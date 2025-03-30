//go:build linux
// +build linux

// eventloop for linux with epoll
package eventloop

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type Task interface {
	GetFD() int               // returns file descriptor
	GetDataToSend() []byte    // data to be written to socket
	Acknowledge([]byte) error // on receiving ack
}

type EventLoop struct {
	events        []unix.EpollEvent                    // event list
	epFd          int                                  // epoll file descriptor
	q             *queue.Queue[Task]                   // task queue
	dump          *collection.ConcurrentMap[int, Task] // task to be dumped after receiving ack
	registeredFDs map[int]bool
}

var eventLoopInstance *EventLoop

// GetEventLoop returns the singleton instance of the EventLoop.
// It initializes the EventLoop if it hasn't been created yet.
func GetEventLoop() (*EventLoop, error) {
	if eventLoopInstance != nil {
		return eventLoopInstance, nil
	}
	eventLoopInstance, err := newEventLoop()
	if err != nil {
		return nil, err
	}
	return eventLoopInstance, nil
}

// newEventLoop creates EventLoop instance
func newEventLoop() (*EventLoop, error) {
	epFd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("error creating epoll instance: %v", err)
	}
	eventLoop := EventLoop{
		q:             queue.NewQueue[Task](),
		epFd:          epFd,
		events:        make([]unix.EpollEvent, 10),
		dump:          collection.NewConcurrentMap[int, Task](),
		registeredFDs: make(map[int]bool),
	}
	return &eventLoop, nil
}

// AddTask pushes Task to task queue
func (t *EventLoop) AddTask(task Task) {
	t.q.Push(task)
}

// SetNonBlocking sets the file descriptor to non-blocking mode
func (t *EventLoop) SetNonBlocking(fd int) error {
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		return fmt.Errorf("failed to get flags: %w", err)
	}
	_, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags|unix.O_NONBLOCK)
	if err != nil {
		return fmt.Errorf("failed to set non-blocking mode: %w", err)
	}
	return nil
}

// SetBlocking sets the file descriptor to blocking mode
func (t *EventLoop) SetBlocking(fd int) error {
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		return fmt.Errorf("failed to get flags: %w", err)
	}
	_, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags&^unix.O_NONBLOCK)
	if err != nil {
		return fmt.Errorf("failed to set blocking mode: %w", err)
	}
	return nil
}

// StartExecLoop starts a goroutine that continuously polls tasks from the queue,
// sets the corresponding file descriptors to blocking mode, registers them with epoll,
// and writes the task's data to the file descriptor.
func (t *EventLoop) StartExecLoop() {
	go func() {
		for {
			if t.q.IsEmpty() {
				continue
			}
			task := t.q.Poll()
			log.Infof("Polling : %v", task)
			fd := task.GetFD()
			t.dump.Set(fd, task)

			if err := t.SetBlocking(fd); err != nil {
				log.Errorf("error setting blocking mode: %v", err)
				continue
			}

			if err := t.registerEvent(fd); err != nil {
				log.Errorf("failed to register event: %v", err)
			}

			buff := task.GetDataToSend()

			_, err := unix.Write(fd, buff)
			if err != nil {
				log.Errorf("write error: %v\n", err)
			}
		}
	}()
}

// registerEvent registers a file descriptor (fd) with the epoll for read events.
func (t *EventLoop) registerEvent(fd int) error {
	if t.registeredFDs[fd] {
		return nil
	}

	event := unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLET,
		Fd:     int32(fd),
	}

	if err := unix.EpollCtl(t.epFd, unix.EPOLL_CTL_ADD, fd, &event); err != nil {
		unix.Close(fd)
		return fmt.Errorf("error adding client to epoll: %v", err)
	}

	t.registeredFDs[fd] = true
	return nil
}

// StartAckLoop starts a goroutine that continuously listens for events from epoll.
//
// Upon receiving an event, it reads the message length, then reads the complete message
// from the corresponding file descriptor and acknowledges the task using the associated handler.
func (t *EventLoop) StartAckLoop() {
	go func() {
		for {
			n, err := unix.EpollWait(t.epFd, t.events, -1)
			if err != nil {
				log.Errorf("Error in EpollWait: %v", err)
				break
			}
			for i := 0; i < n; i++ {
				if err := t.handleEvent(t.events[i]); err != nil {
					log.Errorf("failed to handle event: %v", err)
				}
			}
		}
	}()
}

// handleEvent processes an epoll event for a given file descriptor (fd).
func (t *EventLoop) handleEvent(event unix.EpollEvent) error {
	fd := int(event.Fd)
	lengthBytes := make([]byte, 4)
	_, err := unix.Read(fd, lengthBytes)
	if err != nil {
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return nil
		}
		unix.Close(fd)
		return fmt.Errorf("unexpected read failure: %v", err)
	}
	totalLength := binary.BigEndian.Uint32(lengthBytes)

	var buffer []byte
	for {
		length := totalLength - uint32(len(buffer))
		buf := make([]byte, length)
		n, err := unix.Read(fd, buf)
		if err != nil {
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				return nil
			}
			unix.Close(fd)
			return fmt.Errorf("unexpected read failure: %v", err)
		}
		if n == 0 {
			unix.Close(fd)
			return fmt.Errorf("client disconnected: %v", fd)
		}
		buffer = append(buffer, buf[:n]...)

		if len(buffer) >= int(totalLength) {
			break
		}
	}

	task, _ := t.dump.Get(fd)
	task.Acknowledge(buffer)

	return nil
}

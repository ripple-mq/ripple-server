//go:build darwin
// +build darwin

// eventloop for mac with kqueue
package eventloop

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
)

type Task interface {
	GetFD() int               // returns file descriptor
	GetDataToSend() []byte    // data to be writen to socket
	Acknowledge([]byte) error // on recieveing ack
}

type EventLoop struct {
	events []unix.Kevent_t                      // event list
	kqFd   int                                  // kqueue file descriptor
	q      *queue.Queue[Task]                   // task queue
	dump   *collection.ConcurrentMap[int, Task] // task to be dumped after receciving ack
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
	kqId, err := unix.Kqueue()
	if err != nil {
		return nil, fmt.Errorf("error creating epoll instance: %v", err)
	}
	eventLoop := EventLoop{
		q:      queue.NewQueue[Task](),
		kqFd:   kqId,
		events: make([]unix.Kevent_t, config.Conf.Gossip.EventLoop.Kqueue.Event_buffer_size),
		dump:   collection.NewConcurrentMap[int, Task](),
	}
	return &eventLoop, nil
}

// AddTask pushes Task to task queue
func (t *EventLoop) AddTask(task Task) error {
	if t.q.Size() == int(config.Conf.Gossip.EventLoop.Task_queue_buffer_size) {
		return fmt.Errorf("task buffer overflow")
	}
	t.q.Push(task)
	return nil
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
// sets the corresponding file descriptors to blocking mode, registers them with kqueue,
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
				log.Errorf("error setting non-blocking mode: %v", err)
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

// registerEvent registers a file descriptor (fd) with the kqueue for read events.
// It adds the fd to the kqueue with EV_ADD and EV_ENABLE flags to start monitoring.
// If registration fails, the fd is closed, and an error is returned.
func (t *EventLoop) registerEvent(fd int) error {
	clientEvent := unix.Kevent_t{
		Ident:  uint64(fd),
		Filter: unix.EVFILT_READ,
		Flags:  unix.EV_ADD | unix.EV_ENABLE,
	}

	if _, err := unix.Kevent(t.kqFd, []unix.Kevent_t{clientEvent}, nil, nil); err != nil {
		unix.Close(fd)
		return fmt.Errorf("error adding client to kqueue: %v", err)
	}

	return nil
}

// StartAckLoop starts a goroutine that continuously listens for events from kqueue.
//
// Upon receiving an event, it reads the message length, then reads the complete message
// from the corresponding file descriptor and acknowledges the task using the associated handler.
func (t *EventLoop) StartAckLoop() {
	x := 0
	go func() {
		for {
			n, err := unix.Kevent(t.kqFd, nil, t.events, nil)
			x++
			log.Infof("Kevent: %d", x)
			if err != nil {
				log.Infof("Error in Kevent: %v", err)
				break
			}
			for i := range n {
				if err := t.handleEvent(t.events[i]); err != nil {
					log.Errorf("failed to handle event: %v", err)
				}
			}
		}
	}()
}

// handleEvent processes a kqueue event for a given file descriptor (fd).
//
// It reads the message length, then reads the complete message from the fd.
// If the client disconnects or an error occurs, the fd is closed, and an error is returned.
// Upon successful read, the message is passed to the task's Acknowledge method.
func (t *EventLoop) handleEvent(event unix.Kevent_t) error {
	fd := int(event.Ident)
	lengthBytes := make([]byte, 4)
	_, _ = unix.Read(fd, lengthBytes)
	totalLength := binary.BigEndian.Uint32(lengthBytes)

	var buffer []byte
	// trying to read full `totalLength` bytes of data with blocking read
	for {
		length := totalLength - uint32(len(buffer))
		buf := make([]byte, length)
		n, err := unix.Read(fd, buf)
		if err != nil {
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				return fmt.Errorf("resource error while reading: %v", err)
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

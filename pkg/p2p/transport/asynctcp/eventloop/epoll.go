//go:build linux
// +build linux

package eventloop

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	ic "github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/comm"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/utils"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/comm"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

const (
	EPOLLIN       = 0x001
	EPOLLOUT      = 0x004
	EPOLLET       = 0x80000000
	EPOLL_CTL_ADD = 1
	EPOLL_CTL_MOD = 3
	EPOLL_CTL_DEL = 2
)

type Server struct {
	epollFd     int
	listenerFd  int
	tcpListener net.Listener
	listeners   *collection.ConcurrentMap[string, *ic.Subscriber]
	clients     *collection.ConcurrentMap[int, string]

	Decoder          encoder.Decoder
	mu               sync.Mutex
	isLoopRunning    *collection.ConcurrentValue[bool]
	shutdownSignalCh chan struct{}
}

var serverInstance *Server
var instanceCreationLock = sync.Mutex{}
var instanceAccessLock = sync.Mutex{}

// GetServer returns a singleton instance of the Server for the given address.
// It initializes the server only once and reuses the existing instance on subsequent calls.
//
// Note: Reinstantiation is only possible after stutting down existing one
func GetServer(addr string) (*Server, error) {
	instanceAccessLock.Lock()
	defer instanceAccessLock.Unlock()

	if !instanceCreationLock.TryLock() {
		return serverInstance, nil
	}

	sv, err := newServer(addr)
	serverInstance = sv
	if err != nil {
		instanceCreationLock.Unlock()
		return nil, err
	}

	return serverInstance, nil
}

// newServer initializes a TCP server, sets up a kqueue for event notification,
// and registers the listener for read events. It returns the server instance or an error if any step fails.
func newServer(addr string) (*Server, error) {
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println("ERR, ", err)
		return nil, err
	}

	// Set the Socket operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		fmt.Println("ERR, ", err)
		return nil, err
	}

	// Bind the IP and the port
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Println("ERR, ", err)
		return nil, err
	}
	p, _ := strconv.Atoi(portStr)
	ip4 := net.ParseIP(host)
	fmt.Println("Listening at, ", ip4, p)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: p,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		fmt.Println("ERR, ", err)
		return nil, err
	}

	// Start listening
	if err = syscall.Listen(serverFD, 1000); err != nil {
		fmt.Println("ERR, ", err)
		return nil, err
	}

	// AsyncIO starts here!!

	// creating EPOLL instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}

	// Specify the events we want to get hints about
	// and set the socket on which
	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// Listen to read events on the Server itself
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		fmt.Println("ERR, ", err)
		return nil, err
	}

	return &Server{
		epollFd:          epollFD,
		listenerFd:       serverFD,
		clients:          collection.NewConcurrentMap[int, string](),
		Decoder:          encoder.GOBDecoder{},
		listeners:        collection.NewConcurrentMap[string, *ic.Subscriber](),
		mu:               sync.Mutex{},
		isLoopRunning:    collection.NewConcurrentValue(false),
		shutdownSignalCh: make(chan struct{}),
	}, nil
}

// Subscribe adds a subscriber to the server's listener map with the given ID.
// It allows the server to push messages to respective subscriber channel.
func (t *Server) Subscribe(id string, subscriber *ic.Subscriber) {
	t.listeners.Set(id, subscriber)
}

// UnSubscribe removes the subscriber with the given ID from the server's listener map.
// This stops the server from sending messages to that subscriber.
func (t *Server) UnSubscribe(id string) {
	t.listeners.Delete(id)
}

// Run starts the server event loop, monitoring for I/O events using kqueue.
// It handles new incoming connections and reads data from existing connections.
// Errors during event processing, connection acceptance, or data reading are logged.
func (t *Server) Run() {
	if t.isLoopRunning.Get() {
		return
	}
	t.isLoopRunning.Set(true)
	defer t.Clean()
	log.Info("Started Eventloop...")
	events := make([]syscall.EpollEvent, 10)
	for {
		select {
		case <-t.shutdownSignalCh:
			return
		default:
			n, err := syscall.EpollWait(t.epollFd, events[:], -1)
			if err != nil {
				continue
			}

			for i := range n {
				event := events[i]
				fd := int(event.Fd)

				if fd == t.listenerFd {
					if err := t.Accept(); err != nil {
						log.Errorf("failed to accept new connection: %v", err)
					}
				} else {
					if err := t.Read(fd); err != nil {
						log.Errorf("failed to read message: %v", err)
					}
				}
			}
		}
	}
}

// Accept handles new incoming connections on the server's listener socket.
// It accepts the connection, adds it to the list of active clients, and registers it with kqueue for read events.
// Any errors during acceptance, client registration, or kqueue addition are logged and returned.
func (t *Server) Accept() error {
	connFd, sa, err := syscall.Accept(t.listenerFd)
	if err != nil {
		return fmt.Errorf("accept error: %v", err)
	}

	syscall.SetNonblock(t.listenerFd, true)
	addr := utils.SockaddrToString(sa)
	t.clients.Set(connFd, addr)

	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(connFd),
	}

	if err := syscall.EpollCtl(t.epollFd, EPOLL_CTL_ADD, connFd, &event); err != nil {
		fmt.Println("Error adding connection to epoll:", err)
		syscall.Close(connFd)
		return fmt.Errorf("error adding connection to epoll: %v", err)
	}
	fmt.Printf("Accepted connection from fd %d\n", connFd)
	return nil
}

// Send sends data to a connected client identified by its address.
// If the connection does not exist, it establishes a new TCP connection, registers it with kqueue for read events,
// and sends the data. Errors during connection, registration, or data transmission are logged and returned.
func (t *Server) Send(address string, metadata []byte, data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if connection exists
	keys, values := t.clients.Entries()
	for i := range len(keys) {
		if values[i] == address {
			_, err := t.write(keys[i], data)
			return err
		}
	}

	// If not connected, initiate a new connection
	// connect to a remote server over TCP, retrieves the connection's file descriptor,
	// and register it with kqueue to monitor for read events. Send initial data over the connection.

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", address, err)
	}

	connFile, err := conn.(*net.TCPConn).File()
	if err != nil {
		return fmt.Errorf("failed to get file descriptor: %v", err)
	}

	fd := int(connFile.Fd())

	// Register the new connection with epoll for both read and write events
	event := syscall.EpollEvent{
		Events: uint32(syscall.EPOLLIN) | syscall.EPOLLOUT,
		Fd:     int32(fd),
	}

	syscall.SetNonblock(fd, true)
	if err := syscall.EpollCtl(t.epollFd, syscall.EPOLL_CTL_ADD, fd, &event); err != nil {
		conn.Close()
		return fmt.Errorf("failed to register new connection: %v", err)
	}

	// Add the connection to the clients map
	t.clients.Set(fd, address)

	// Send the data
	_, _ = t.write(fd, metadata)
	_, err = t.write(fd, data)
	return err
}

// Blocking write
func (t *Server) write(fd int, data []byte) (int, error) {
	written := 0
	for written < len(data) {
		n, err := syscall.Write(fd, data[written:])
		if n > 0 {
			written += n
		} else if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
			time.Sleep(10 * time.Millisecond)
			continue
		} else {
			return 0, fmt.Errorf("failed to write data: %v", err)
		}
	}
	return written, nil
}

// Read handles incoming data from a client connection identified by the kqueue event.
// It reads the message length, then the actual data, decodes it, and forwards it to the appropriate subscriber.
// If the subscriber hasn't been greeted yet, it sends a greeting; otherwise, it pushes the message.
// Errors during reading or processing lead to client removal and error reporting.
func (t *Server) Read(fd int) error {
	lengthBytes := make([]byte, 4)
	_, _ = syscall.Read(fd, lengthBytes)
	length := binary.BigEndian.Uint32(lengthBytes)

	buf := make([]byte, length)
	n, err := t.read(fd, buf)
	if err != nil {
		t.removeClient(fd)
		return fmt.Errorf("read error: %v", err)
	}

	if n > 0 {
		payload := t.decodeToPayload(buf[:n])
		subscriber, err := t.listeners.Get(payload.ID)
		if err != nil {
			// Invalid Id might indicate unsubscribed servers, so dropping connection
			t.removeClient(fd)
			return fmt.Errorf("no listeners found with ID %s, error = %v", payload.ID, err)
		}

		addr, _ := t.clients.Get(fd)
		if !subscriber.GreetStatus.Get() {
			fmt.Println("grreting metadata")
			subscriber.Greet(ic.Message{RemoteAddr: addr, RemoteID: payload.FromServerID, Payload: payload.Data})
			subscriber.GreetStatus.Set(true)
		} else {
			subscriber.Push(ic.Message{RemoteAddr: addr, RemoteID: payload.FromServerID, Payload: payload.Data})
		}
	} else {
		t.removeClient(fd)
	}
	return nil
}

// read completely
func (t *Server) read(fd int, buffer []byte) (int, error) {
	totalRead := 0
	for totalRead < len(buffer) {
		n, err := syscall.Read(fd, buffer[totalRead:])
		if err != nil {
			fmt.Println("ERR, ", err)
			return totalRead, err
		}
		if n == 0 {
			break
		}
		totalRead += n
	}
	return totalRead, nil
}

// removeClient closes the client connection and removes it from the active client list.
func (t *Server) removeClient(fd int) {
	if _, err := t.clients.Get(fd); err == nil {

		t.clients.Delete(fd)
		if err := syscall.EpollCtl(t.epollFd, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
			log.Warnf("Failed to remove FD %d from epoll: %v", fd, err)
		}
		er := syscall.Close(fd)
		fmt.Printf("Client %d disconnected : %v\n", fd, er)
	}
}

// decodeToPayload decodes raw data into a comm.Payload structure using the server's decoder.
func (t *Server) decodeToPayload(data []byte) comm.Payload {
	var msg comm.Payload
	err := t.Decoder.Decode(bytes.NewBuffer(data), &msg)
	if err != nil {
		fmt.Println("Decode error, ", err)
	}
	return msg
}

func (t *Server) Clean() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Println("CLEANING STARTED >>>>>>>>>>>>>> ")
	// release lock to allow creating new Server instance
	instanceCreationLock.Unlock()

	if err := t.tcpListener.Close(); err != nil {
		log.Errorf("failed to close listener at %s, error= %v", t.tcpListener.Addr(), err)
	}
	// Close all client connections
	for fd := range t.clients.Keys() {
		err := syscall.Close(fd)
		if err != nil {
			log.Errorf("Error closing client connection %d: %v", fd, err)
		} else {
			log.Infof("Closed client connection %d", fd)
		}
	}

	// Close the listener socket
	if err := syscall.Close(t.listenerFd); err != nil {
		log.Errorf("Error closing listener: %v", err)
	} else {
		log.Info("Listener socket closed.")
	}

	// Close the kqueue
	if err := syscall.Close(t.epollFd); err != nil {
		log.Errorf("Error closing kqueue: %v", err)
	} else {
		log.Info("Kqueue closed.")
	}

	// Clear subscribers
	return nil
}

// Stop stops server, blocking in nature
func (t *Server) Stop() {
	syscall.Close(t.listenerFd)
}

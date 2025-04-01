//go:build linux
// +build linux

package eventloop

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"syscall"

	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	ic "github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/comm"
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
	epollFd    int
	listenerFd int
	Decoder    encoder.Decoder
	clients    map[int]string
	listeners  *collection.ConcurrentMap[string, *ic.Subscriber]
	mu         sync.Mutex
}

var serverInstance *Server

// GetServer returns a singleton instance of the Server, creating it if it doesn't exist.
func GetServer(addr string) (*Server, error) {
	if serverInstance != nil {
		return serverInstance, nil
	}
	sv, err := newServer(addr)
	serverInstance = sv
	if err != nil {
		return nil, err
	}
	return serverInstance, nil
}

// newServer initializes a TCP server, sets up a kqueue for event notification,
// and registers the listener for read events. It returns the server instance or an error if any step fails.
func newServer(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error starting server: %v", err)
	}

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, fmt.Errorf("failed to assert listener to *net.TCPListener")
	}

	listenerFile, err := tcpListener.File()
	if err != nil {
		return nil, fmt.Errorf("error getting listener file descriptor: %v", err)
	}

	fd := int(listenerFile.Fd())
	epollFd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("error creating epoll: %v", err)
	}

	event := syscall.EpollEvent{
		Events: uint32(syscall.EPOLLIN) | EPOLLET,
		Fd:     int32(fd),
	}

	if err := syscall.EpollCtl(epollFd, EPOLL_CTL_ADD, fd, &event); err != nil {
		return nil, fmt.Errorf("error adding listener to epoll: %v", err)
	}

	return &Server{
		epollFd:    epollFd,
		listenerFd: fd,
		clients:    make(map[int]string),
		Decoder:    encoder.GOBDecoder{},
		listeners:  collection.NewConcurrentMap[string, *ic.Subscriber](),
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
	events := make([]syscall.EpollEvent, 10)
	for {
		n, err := syscall.EpollWait(t.epollFd, events, -1)
		if err != nil {
			fmt.Println("Error waiting for events:", err)
			break
		}

		for i := 0; i < n; i++ {
			event := events[i]
			fd := int(event.Fd)

			if fd == t.listenerFd {
				t.Accept()
			} else {
				t.Read(fd)
			}
		}
	}
}

// Accept handles new incoming connections on the server's listener socket.
// It accepts the connection, adds it to the list of active clients, and registers it with kqueue for read events.
// Any errors during acceptance, client registration, or kqueue addition are logged and returned.
func (t *Server) Accept() error {
	connFd, _, err := syscall.Accept(t.listenerFd)
	if err != nil {
		fmt.Println("Accept error:", err)
		return fmt.Errorf("accept error: %v", err)
	}

	t.mu.Lock()
	t.clients[connFd] = fmt.Sprintf("fd_%d", connFd)
	t.mu.Unlock()

	event := syscall.EpollEvent{
		Events: uint32(syscall.EPOLLIN) | EPOLLET,
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
func (t *Server) Send(address string, data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if connection exists
	for fd, addr := range t.clients {
		if addr == address {
			_, err := syscall.Write(fd, data)
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
		Events: uint32(syscall.EPOLLIN) | syscall.EPOLLOUT | EPOLLET,
		Fd:     int32(fd),
	}

	if err := syscall.EpollCtl(t.epollFd, syscall.EPOLL_CTL_ADD, fd, &event); err != nil {
		conn.Close()
		return fmt.Errorf("failed to register new connection: %v", err)
	}

	// Add the connection to the clients map
	t.clients[fd] = address

	// Send the data
	_, err = syscall.Write(fd, data)
	return err
}

// Read handles incoming data from a client connection identified by the kqueue event.
// It reads the message length, then the actual data, decodes it, and forwards it to the appropriate subscriber.
// If the subscriber hasn't been greeted yet, it sends a greeting; otherwise, it pushes the message.
// Errors during reading or processing lead to client removal and error reporting.
func (t *Server) Read(fd int) {
	buf := make([]byte, 1024)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		fmt.Println("Read error:", err)
		t.removeClient(fd)
		return
	}

	if n > 0 {
		payload := t.decodeToPayload(buf[:n])
		subscriber, err := t.listeners.Get(payload.ID)
		if err != nil {
			fmt.Println("Error: ", err, payload)
			return
		}

		if !subscriber.GreetStatus.Get() {
			subscriber.Greet(ic.Message{RemoteAddr: t.clients[fd], Payload: payload.Data})
			subscriber.GreetStatus.Set(true)
		} else {
			subscriber.Push(ic.Message{RemoteAddr: t.clients[fd], Payload: payload.Data})
		}
	} else {
		t.removeClient(fd)
	}
}

// removeClient closes the client connection and removes it from the active client list.
func (t *Server) removeClient(fd int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.clients[fd]; exists {
		syscall.Close(fd)
		delete(t.clients, fd)
		fmt.Printf("Client %d disconnected\n", fd)
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

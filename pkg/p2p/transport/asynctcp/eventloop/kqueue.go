//go:build darwin
// +build darwin

package eventloop

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	ic "github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/comm"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/utils"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/comm"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

const (
	EVFILT_READ  = -1
	EVFILT_WRITE = -2
	EV_ADD       = 0x0001
	EV_ENABLE    = 0x0000
	EV_ONESHOT   = 0x0002
)

type Server struct {
	kq         int
	listenerFd int
	Decoder    encoder.Decoder
	clients    map[int]string
	listeners  *collection.ConcurrentMap[string, *ic.Subscriber]
	mu         sync.Mutex
}

var serverInstance *Server

// GetServer returns a singleton instance of the Server for the given address.
// It initializes the server only once and reuses the existing instance on subsequent calls.
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
	kq, err := syscall.Kqueue()
	if err != nil {
		return nil, fmt.Errorf("error creating kqueue: %v", err)
	}

	kev := syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: EVFILT_READ,
		Flags:  EV_ADD | EV_ENABLE,
	}

	if _, err := syscall.Kevent(kq, []syscall.Kevent_t{kev}, nil, nil); err != nil {
		return nil, fmt.Errorf("error registering listener with kqueue: %v", err)
	}

	return &Server{
		kq:         kq,
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
	events := make([]syscall.Kevent_t, 10)
	for {
		n, err := syscall.Kevent(t.kq, nil, events, nil)
		if err != nil {
			log.Errorf("error waiting for events: %v", err)
			break
		}

		for i := range n {
			event := events[i]
			if event.Flags&syscall.EV_ERROR != 0 {
				log.Errorf("event error: %v", event)
				continue
			}

			if int(event.Ident) == t.listenerFd {
				if err := t.Accept(); err != nil {
					log.Errorf("failed to accept new connection: %v", err)
				}
			} else {
				if err := t.Read(event); err != nil {
					log.Errorf("failed to read message: %v", err)
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
		fmt.Println("Accept error:", err)
		return fmt.Errorf("accept error: %v", err)
	}

	addr := utils.SockaddrToString(sa)
	t.mu.Lock()
	t.clients[connFd] = addr
	t.mu.Unlock()

	kev := syscall.Kevent_t{
		Ident:  uint64(connFd),
		Filter: EVFILT_READ,
		Flags:  EV_ADD | EV_ENABLE,
	}
	if _, err := syscall.Kevent(t.kq, []syscall.Kevent_t{kev}, nil, nil); err != nil {
		fmt.Println("Error adding connection to kqueue:", err)
		syscall.Close(connFd)
		return fmt.Errorf("error adding connection to kqueue: %v", err)
	}
	fmt.Printf("Accepted connection from %s\n", addr)
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
	kev := syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: EVFILT_READ,
		Flags:  EV_ADD | EV_ENABLE,
	}

	if _, err := syscall.Kevent(t.kq, []syscall.Kevent_t{kev}, nil, nil); err != nil {
		conn.Close()
		return fmt.Errorf("failed to register new connection: %v", err)
	}

	t.clients[fd] = address
	_, err = syscall.Write(fd, data)
	return err
}

// Read handles incoming data from a client connection identified by the kqueue event.
// It reads the message length, then the actual data, decodes it, and forwards it to the appropriate subscriber.
// If the subscriber hasn't been greeted yet, it sends a greeting; otherwise, it pushes the message.
// Errors during reading or processing lead to client removal and error reporting.
func (t *Server) Read(event syscall.Kevent_t) error {
	lengthBytes := make([]byte, 4)
	_, _ = syscall.Read(int(event.Ident), lengthBytes)
	length := binary.BigEndian.Uint32(lengthBytes)

	buf := make([]byte, length)
	n, err := syscall.Read(int(event.Ident), buf)
	if err != nil {
		t.removeClient(int(event.Ident))
		return fmt.Errorf("read error: %v", err)
	}
	if n > 0 {
		payload := t.decodeToPayload(buf[:n])
		subscriber, err := t.listeners.Get(payload.ID)
		if err != nil {
			return err
		}

		if !subscriber.GreetStatus.Get() {
			subscriber.Greet(ic.Message{RemoteAddr: t.clients[int(event.Ident)], Payload: payload.Data})
			subscriber.GreetStatus.Set(true)
		} else {
			subscriber.Push(ic.Message{RemoteAddr: t.clients[int(event.Ident)], Payload: payload.Data})
		}
	} else {
		t.removeClient(int(event.Ident))
	}

	return nil
}

// removeClient closes the client connection and removes it from the active client list.
func (s *Server) removeClient(fd int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[fd]; exists {
		syscall.Close(fd)
		delete(s.clients, fd)
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

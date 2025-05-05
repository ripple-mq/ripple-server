//go:build newimpl
// +build newimpl

package eventloop

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/peer"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/comm"
	ic "github.com/ripple-mq/ripple-server/pkg/server/asynctcp/comm"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

// Server represents a network server utilizing kqueue for I/O multiplexing, handling connections and events.
// It manages incoming connections, active clients, listeners, and event processing.
type Server struct {
	listener   net.Listener
	listenAddr net.Addr
	clients    *collection.ConcurrentMap[int, string]            // Map of active clients
	listeners  *collection.ConcurrentMap[string, *ic.Subscriber] // Map of active listeners

	Decoder          encoder.Decoder                   // Decoder for incoming messages
	mu               *sync.RWMutex                     // Mutex for synchronization
	isLoopRunning    *collection.ConcurrentValue[bool] // Flag indicating if the event loop is running
	shutdownSignalCh chan struct{}                     // Channel for shutdown signal
	wg               sync.WaitGroup
	PeersMap         map[string]peer.Peer // Map storing active peers by address
}

var serverInstance *Server
var instanceCreationLock = sync.Mutex{}
var instanceAccessLock = sync.Mutex{}
var writeLock = sync.Mutex{}
var readLock = sync.Mutex{}

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
	log.Info("Creating Eventloop V2")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener:         listener,
		listenAddr:       listener.Addr(),
		clients:          collection.NewConcurrentMap[int, string](),
		Decoder:          encoder.GOBDecoder{},
		listeners:        collection.NewConcurrentMap[string, *ic.Subscriber](),
		isLoopRunning:    collection.NewConcurrentValue(false),
		mu:               &sync.RWMutex{},
		shutdownSignalCh: make(chan struct{}),
		wg:               sync.WaitGroup{},
		PeersMap:         make(map[string]peer.Peer),
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

	defer func() {
		log.Infof("Shutting down server: %s", t.listenAddr)
		t.wg.Done()
	}()

	t.wg.Add(1)
	for {

		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.Warnf("failed to establish connection with %s", t.listenAddr.String())
		}

		t.addConnection(conn)
		go t.handleConnection(conn)

	}
}

// handleConnection processes incoming data from a network connection, handling message length and payload.
// It calls the OnAcceptingConn function on the first message and enqueues subsequent messages to the incoming queue.
func (t *Server) handleConnection(conn net.Conn) {
	isOnConnExecuted := false
	defer func() {
		err := t.dropConnection(conn.RemoteAddr().String())
		if err != nil {
			log.Debugf("Error while dropping connection: %v", err)
		}
		log.Debugf("Dropping connection: %s\n", conn.RemoteAddr())
	}()

	for {
		// first 4 bytes are used for message length
		lengthBytes := make([]byte, 4)
		_, err := conn.Read(lengthBytes)
		length := binary.BigEndian.Uint32(lengthBytes)

		if err == io.EOF {
			return
		}

		var buffer = make([]byte, length)
		n, _ := conn.Read(buffer)

		if !isOnConnExecuted {
			// t.OnAcceptingConn(conn, buffer[:n])
			isOnConnExecuted = true
			continue
		}
		payload := t.decodeToPayload(buffer[:n])
		subscriber, err := t.listeners.Get(payload.ID)
		if err != nil {
			// Invalid Id might indicate unsubscribed servers, so dropping connection
			t.dropConnection(conn.RemoteAddr().String())

		}
		log.Debugf("Recieved message of length %d from %s", n, conn.RemoteAddr().String())
		// t.IncommingMsgQueue <- Message{RemoteAddr: conn.RemoteAddr().String(), Payload: buffer[:n]}
		subscriber.Push(ic.Message{RemoteAddr: conn.RemoteAddr().String(), RemoteID: payload.FromServerID, Payload: payload.Data})

	}
}

// getConnection retrieves an existing connection for the given address from the PeersMap.
// Returns the peer associated with the address or an error if no connection exists.
func (t *Server) getConnection(addr string) (peer.Peer, error) {
	// t.mu.RLock()
	// defer t.mu.RUnlock()
	tcpAddr := t.buildAddress(addr)
	if conn := t.PeersMap[tcpAddr]; conn != nil {
		return conn, nil
	}
	return nil, fmt.Errorf("failed to get connection")
}

// addConnection adds a new connection to the PeersMap and returns the corresponding peer.
func (t *Server) addConnection(conn net.Conn) peer.Peer {
	// t.mu.Lock()
	x := conn.RemoteAddr().String()
	t.PeersMap[x] = &peer.TCPPeer{Conn: conn}
	res := t.PeersMap[x]
	// t.mu.Unlock()
	return res
}

// dropConnection removes a connection for the given address from the PeersMap and closes it.
// Returns an error if no connection exists for the address.
func (t *Server) dropConnection(addr string) error {
	// t.mu.Lock()
	// defer t.mu.Unlock()
	ResolveTCPAddr := t.buildAddress(addr)
	if conn := t.PeersMap[ResolveTCPAddr]; conn != nil {
		delete(t.PeersMap, conn.GetAddress().String())
		return conn.Close()
	}
	return fmt.Errorf("no existing connections found")
}

func (t *Server) buildAddress(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr

}

// Send sends data to a connected client identified by its address.
// If the connection does not exist, it establishes a new TCP connection, registers it with kqueue for read events,
// and sends the data. Errors during connection, registration, or data transmission are logged and returned.
func (t *Server) Send(address string, metadata []byte, data []byte) error {
	// t.mu.Lock()
	// defer t.mu.Unlock()

	peerNode, err := t.dial(address, metadata)
	if err != nil {
		return err
	}
	t.write(peerNode, data)
	return nil
}

// dial attempts to establish a connection to the given address and returns the associated peer node.
//   - If a connection already exists, it returns the existing peer; otherwise, it establishes a new one.
//   - Based on config, it can start listening on same connection.
func (t *Server) dial(addr string, metadata []byte) (peer.Peer, error) {
	peerNode, _ := t.getConnection(addr)
	if peerNode != nil {
		return peerNode, nil
	}
	conn, err := net.Dial(t.listenAddr.Network(), addr)
	peerNode = t.addConnection(conn)

	t.write(peerNode, metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection with %s", addr)
	}

	return peerNode, nil
}

// write encodes the given data and sends it to the specified peer node,
// first sending the length of the message followed by the message itself.
func (t *Server) write(peerNode peer.Peer, data []byte) {

	_, err := peerNode.GetConnection().Write(data)
	if err != nil {
		log.Errorf("failed to send data: %v", err)
	}
}

// decodeToPayload decodes raw data into a comm.Payload structure using the server's decoder.
func (t *Server) decodeToPayload(data []byte) comm.Payload {
	var msg comm.Payload
	err := t.Decoder.Decode(bytes.NewBuffer(data), &msg)
	if err != nil {
		log.Errorf("Decode error, %v", err)
	}
	return msg
}

// Clean gracefully shuts down the server, closing all client connections,
// the listener socket, and the kqueue.
//
// Depricated: Does nothing
func (t *Server) Clean() error {
	return nil
}

// Stop stops server, blocking in nature
func (t *Server) Stop() {
	// t.shutdownSignalCh <- struct{}{}
}

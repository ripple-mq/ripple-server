package tcp

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/charmbracelet/log"

	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/peer"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/comm"
)

type Message struct {
	RemoteAddr string
	Payload    []byte
}

type TransportOpts struct {
	ShouldClientHandleConn bool
	Ack                    bool
}

// Transport provides an implementation for handling network connections, message encoding,
// and managing peers for communication. It supports connection handling, message processing,
// and acknowledgment management.
type Transport struct {
	ListenAddr             net.Addr                            // The address the transport listens on
	IncommingMsgQueue      chan Message                        // Queue for incoming messages
	Encoder                encoder.Encoder                     // Encoder used for encoding messages
	OnAcceptingConn        func(conn net.Conn, message []byte) // Callback for handling accepted connections
	ShouldClientHandleConn bool                                // Flag to determine if the client should handle connections
	Ack                    bool                                // Flag indicating if acknowledgment is enabled

	mu       *sync.RWMutex        // Mutex for synchronizing access to shared resources
	PeersMap map[string]peer.Peer // Map storing active peers by address
	listener net.Listener         // Listener for incoming connections
	wg       sync.WaitGroup       // WaitGroup to manage goroutine synchronization
}

// NewTransport creates a new Transport instance with the specified address, connection handler, and options.
// It returns the initialized Transport and any error encountered during initialization.
func NewTransport(addr string, OnAcceptingConn func(conn net.Conn, message []byte), opts ...TransportOpts) (*Transport, error) {
	defaultOpts := TransportOpts{ShouldClientHandleConn: true, Ack: false}
	if len(opts) > 0 {
		defaultOpts = opts[0]
	}

	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address type, %v", err)
	}
	return &Transport{

		ListenAddr:             address,
		IncommingMsgQueue:      make(chan Message, 10),
		Encoder:                encoder.GOBEncoder{},
		PeersMap:               make(map[string]peer.Peer),
		wg:                     sync.WaitGroup{},
		mu:                     &sync.RWMutex{},
		OnAcceptingConn:        OnAcceptingConn,
		ShouldClientHandleConn: defaultOpts.ShouldClientHandleConn,
		Ack:                    defaultOpts.Ack,
	}, nil
}

// Listen starts server, accepts new connection
func (t *Transport) Listen() error {
	var err error
	t.listener, err = net.Listen(t.ListenAddr.Network(), t.ListenAddr.String())
	if err != nil {
		return fmt.Errorf("failed to start server, %v", err)
	}
	log.Infof("TCP: Started listening at port %s", t.ListenAddr)
	go t.connectionLoop()
	return nil
}

// Stop stops accepting new connections
// Still it can send & recieve messages on existing connection
func (t *Transport) Stop() error {
	return t.listener.Close()
}

// Send sends message, metadata will be sent only for first connection
func (t *Transport) Send(addr string, metadata any, data any) error {
	return t.send(addr, metadata, data)
}

// SendToAsync wraps data with server id , sends to async server
func (t *Transport) SendToAsync(addr string, id string, metadata any, data any) error {
	var metadataBuf, dataBuf bytes.Buffer
	encoder.GOBEncoder{}.Encode(metadata, &metadataBuf)
	encoder.GOBEncoder{}.Encode(data, &dataBuf)
	metadataPayload := comm.Payload{ID: id, Data: metadataBuf.Bytes()}
	dataPayload := comm.Payload{ID: id, Data: dataBuf.Bytes()}

	return t.Send(addr, metadataPayload, dataPayload)
}

// Close drops existing connection
func (t *Transport) Close(addr string) error {
	return t.dropConnection(addr)
}

// Consume consumes message from message queue
func (t *Transport) Consume(decoder encoder.Decoder, writer any) (string, error) {
	data := <-t.IncommingMsgQueue
	err := decoder.Decode(bytes.NewBuffer(data.Payload), writer)
	if err != nil {
		return "", err
	}
	return data.RemoteAddr, nil
}

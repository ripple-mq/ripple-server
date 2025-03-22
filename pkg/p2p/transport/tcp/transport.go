package tcp

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/charmbracelet/log"

	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/peer"
)

type Message struct {
	RemoteAddr string
	Payload    []byte
}

// Transport implements Transport
type Transport struct {
	ListenAddr        net.Addr
	IncommingMsgQueue chan Message
	Encoder           encoder.Encoder
	OnAcceptingConn   func(conn net.Conn)

	mu       *sync.RWMutex
	PeersMap map[string]peer.Peer
	listener net.Listener
	wg       sync.WaitGroup
}

func NewTransport(addr string) (*Transport, error) {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address type, %v", err)
	}
	return &Transport{

		ListenAddr:        address,
		IncommingMsgQueue: make(chan Message, 10),
		Encoder:           encoder.GOBEncoder{},
		PeersMap:          make(map[string]peer.Peer),
		wg:                sync.WaitGroup{},
		mu:                &sync.RWMutex{},
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

// Send sends message
func (t *Transport) Send(addr string, data any) error {
	return t.send(addr, data)
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

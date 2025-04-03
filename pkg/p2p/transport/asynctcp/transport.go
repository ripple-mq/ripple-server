package asynctcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/comm"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/eventloop"
	tcpcomm "github.com/ripple-mq/ripple-server/pkg/p2p/transport/comm"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
)

type Transport struct {
	EventLoop       *eventloop.Server
	ListenAddr      comm.ServerAddr
	Encoder         encoder.Encoder
	subscriber      *comm.Subscriber
	Ack             bool
	OnAcceptingConn func(msg comm.Message)
}

type TransportOpts struct {
	Ack             bool
	OnAcceptingConn func(msg comm.Message)
}

// NewTransport creates a new Transport instance with the given ID and optional TransportOpts.
// It initializes the event loop, subscribes to the server, and sets up the encoder and connection handler.
func NewTransport(id string, opts ...TransportOpts) (*Transport, error) {
	var defaultOpts = TransportOpts{OnAcceptingConn: func(msg comm.Message) {}, Ack: false}

	if len(opts) > 0 {
		defaultOpts = opts[0]
	}

	listenAddr := config.Conf.AsyncTCP.Address
	el, err := eventloop.GetServer(listenAddr)
	if err != nil {
		return nil, err
	}
	subscriber := comm.NewSubscriber(id, defaultOpts.OnAcceptingConn)
	el.Subscribe(id, subscriber)
	return &Transport{
		EventLoop:       el,
		ListenAddr:      comm.ServerAddr{Addr: listenAddr, ID: id},
		subscriber:      subscriber,
		Encoder:         encoder.GOBEncoder{},
		Ack:             defaultOpts.Ack,
		OnAcceptingConn: defaultOpts.OnAcceptingConn,
	}, nil
}

// Listen starts the event loop in a separate goroutine to handle incoming connections.
// It returns immediately without waiting for the event loop to complete.
func (t *Transport) Listen() error {
	go t.EventLoop.Run()
	return nil
}

// Send sends metadata (if provided) and data to the specified address.
// It returns an error if sending either the metadata or data fails.
func (t *Transport) Send(addr string, metadata any, data any) error {
	if metadata != struct{}{} {
		if err := t.write(addr, metadata); err != nil {
			return fmt.Errorf("error sending metadata: %v", err)
		}
	}
	if err := t.write(addr, data); err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}
	return nil
}

func (t *Transport) SendToAsync(addr string, id string, metadata any, data any) error {
	var metadataBuf, dataBuf bytes.Buffer
	encoder.GOBEncoder{}.Encode(metadata, &metadataBuf)
	encoder.GOBEncoder{}.Encode(data, &dataBuf)
	metadataPayload := tcpcomm.Payload{ID: id, Data: metadataBuf.Bytes()}
	dataPayload := tcpcomm.Payload{ID: id, Data: dataBuf.Bytes()}

	return t.Send(addr, metadataPayload, dataPayload)
}

// Stop stops eventloop server. Be carefull while using as it has global impact.
func (t *Transport) Stop() error {
	t.EventLoop.Stop()
	return nil
}

// Consume retrieves the next message from the subscriber, decodes it using the provided decoder,
// and writes the decoded data to the given writer. It returns the remote address of the sender or an error if decoding fails.
func (t *Transport) Consume(decoder encoder.Decoder, writer any, timeout ...<-chan time.Time) (string, error) {
	data, err := t.subscriber.Poll(timeout...)
	if err != nil {
		return "", err
	}
	err = decoder.Decode(bytes.NewBuffer(data.Payload), writer)
	if err != nil {
		return "", err
	}
	return data.RemoteAddr, nil
}

// write encodes the provided data using the transport's encoder,
// sends the length of the encoded data followed by the data itself to the specified address.
// Returns an error if either the length or data transmission fails.
func (t *Transport) write(addr string, data any) error {
	var msg bytes.Buffer
	if err := t.Encoder.Encode(data, &msg); err != nil {
		return err
	}

	length := uint32(len(msg.Bytes()))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	err := t.EventLoop.Send(addr, lengthBytes)
	if err != nil {
		return fmt.Errorf("failed to send length: %v", err)
	}
	err = t.EventLoop.Send(addr, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}
	return nil
}

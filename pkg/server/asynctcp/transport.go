package asynctcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	tcpcomm "github.com/ripple-mq/ripple-server/pkg/p2p/transport/comm"
	"github.com/ripple-mq/ripple-server/pkg/server/asynctcp/comm"
	"github.com/ripple-mq/ripple-server/pkg/server/asynctcp/eventloop"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	"github.com/ripple-mq/ripple-server/pkg/utils/env"
)

// Transport manages network communication, including event loops, message encoding,
// and handling incoming connections with optional acknowledgment.
type Transport struct {
	EventLoop       *eventloop.Server      // The event loop server for handling communication
	ListenAddr      comm.ServerAddr        // The address to listen on for incoming connections
	Encoder         encoder.Encoder        // Encoder used for encoding messages
	subscriber      *comm.Subscriber       // Subscriber for receiving messages
	Ack             bool                   // Flag indicating if acknowledgment is enabled
	OnAcceptingConn func(msg comm.Message) // Callback function for handling accepted connections
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

	addr := fmt.Sprintf("%s:%d", env.Get("ASYNC_TCP_IPv4", ""), config.Conf.AsyncTCP.Port)
	listenAddr := addr
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
	if err := t.write(addr, metadata, data); err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}
	return nil
}

func (t *Transport) SendToAsync(addr string, id string, metadata any, data any) error {
	var metadataBuf, dataBuf bytes.Buffer
	encoder.GOBEncoder{}.Encode(metadata, &metadataBuf)
	encoder.GOBEncoder{}.Encode(data, &dataBuf)
	metadataPayload := tcpcomm.Payload{FromServerID: t.ListenAddr.ID, ID: id, Data: metadataBuf.Bytes()}
	dataPayload := tcpcomm.Payload{FromServerID: t.ListenAddr.ID, ID: id, Data: dataBuf.Bytes()}

	return t.Send(addr, metadataPayload, dataPayload)
}

// Stop stops eventloop server. Be carefull while using as it has global impact.
func (t *Transport) Stop() error {
	t.EventLoop.Stop()
	return nil
}

// Consume retrieves the next message from the subscriber, decodes it using the provided decoder,
// and writes the decoded data to the given writer. It returns the remote address of the sender or an error if decoding fails.
func (t *Transport) Consume(decoder encoder.Decoder, writer any, timeout ...<-chan time.Time) (comm.ServerAddr, error) {
	data, err := t.subscriber.Poll(timeout...)

	var null comm.ServerAddr
	if err != nil {
		return null, err
	}
	err = decoder.Decode(bytes.NewBuffer(data.Payload), writer)
	if err != nil {
		return null, err
	}
	return comm.ServerAddr{Addr: data.RemoteAddr, ID: data.RemoteID}, nil
}

// write encodes the provided data using the transport's encoder,
// sends the length of the encoded data followed by the data itself to the specified address.
// Returns an error if either the length or data transmission fails.
func (t *Transport) write(addr string, metadata any, data any) error {

	// Data
	var dataBuff bytes.Buffer
	if err := t.Encoder.Encode(data, &dataBuff); err != nil {
		return err
	}

	dataLength := uint32(len(dataBuff.Bytes()))
	dataLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLengthBytes, dataLength)
	fullData := append(dataLengthBytes, dataBuff.Bytes()...)

	// Metadata
	var metadatBuff bytes.Buffer

	if err := t.Encoder.Encode(metadata, &metadatBuff); err != nil {
		return err
	}
	metadataLength := uint32(len(metadatBuff.Bytes()))
	metadataLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(metadataLengthBytes, metadataLength)
	fullMetadata := append(metadataLengthBytes, metadatBuff.Bytes()...)

	// sending metadata & data together
	err := t.EventLoop.Send(addr, fullMetadata, fullData)
	if err != nil {
		return fmt.Errorf("failed to send length: %v", err)
	}

	return nil
}

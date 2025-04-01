package server

import (
	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp"
)

type ConsumerServer[T queue.PayloadIF] struct {
	ID     string              // listening address
	server *asynctcp.Transport // Consumer server instance
	q      *queue.Queue[T]     // thread safe message queue
}

// NewConsumerServer creates and returns a new ConsumerServer instance.
//
// It initializes a new server that listens on the given address and uses the provided
// message queue. The server will accept incoming connections and handle consumer requests.
func NewConsumerServer[T queue.PayloadIF](id string, q *queue.Queue[T]) (*ConsumerServer[T], error) {
	server, err := asynctcp.NewTransport(id, asynctcp.TransportOpts{OnAcceptingConn: onAcceptingConsumer})
	if err != nil {
		return nil, err
	}

	return &ConsumerServer[T]{
		ID:     id,
		server: server,
		q:      q,
	}, nil
}

// Listen starts the server and begins accepting data from the message queue.
//
// It initializes the server to listen for incoming connections and then starts
// accepting consume requests from the message queue asynchronously.
func (t *ConsumerServer[T]) Listen() error {
	err := t.server.Listen()
	t.startAcceptingConsumeReq()
	return err
}

// Stop stops listening to new Consumer connections
//
// Note: It still continues to serve existing connections
func (t *ConsumerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

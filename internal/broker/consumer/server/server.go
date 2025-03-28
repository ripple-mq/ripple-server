package server

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
)

type ConsumerServer[T any] struct {
	listenAddr net.Addr        // listening address
	server     *tcp.Transport  // Consumer server instance
	q          *queue.Queue[T] // thread safe message queue
}

// NewConsumerServer creates and returns a new ConsumerServer instance.
//
// It initializes a new server that listens on the given address and uses the provided
// message queue. The server will accept incoming connections and handle consumer requests.
func NewConsumerServer[T any](addr string, q *queue.Queue[T]) (*ConsumerServer[T], error) {
	server, err := tcp.NewTransport(addr, onAcceptingConsumer)
	if err != nil {
		return nil, err
	}

	return &ConsumerServer[T]{
		listenAddr: server.ListenAddr,
		server:     server,
		q:          q,
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

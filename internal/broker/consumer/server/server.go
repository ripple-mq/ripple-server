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

// NewConsumerServer[T] creates Sub server to accept data of type T
//
// Parameters:
//   - addr(string): address to listen at
//   - q(*queue.Queue[T]): message queue
//
// Returns:
//   - *ProducerServer[T]
//   - error
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

// Listen starts Sub server & starts accepting data from message queue
//
// Returns:
//   - error
func (t *ConsumerServer[T]) Listen() error {
	err := t.server.Listen()
	t.startAcceptingConsumeReq()
	return err
}

// Stop stops listening to new Consumer connections
func (t *ConsumerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

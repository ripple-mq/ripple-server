package producer

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
)

type ProducerServer[T any] struct {
	listenAddr net.Addr
	server     *tcp.Transport
	q          *queue.Queue[T]
}

// NewProducerServer[T] creates Pub server to accept data of type T
//
// Parameters:
//   - addr(string): address to listen at
//   - q(*queue.Queue[T]): message queue
//
// Returns:
//   - *ProducerServer[T]
//   - error
func NewProducerServer[T any](addr string, q *queue.Queue[T]) (*ProducerServer[T], error) {
	server, err := tcp.NewTransport(addr, onAcceptingProdcuer)
	if err != nil {
		return nil, err
	}

	return &ProducerServer[T]{
		listenAddr: server.ListenAddr,
		server:     server,
		q:          q,
	}, nil
}

// Listen starts Pub server & starts populating data to message queue
//
// Returns:
//   - error
func (t *ProducerServer[T]) Listen() error {
	err := t.server.Listen()
	t.startPopulatingQueue()
	return err
}

// Stop stops listening to new Producer connections
func (t *ProducerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

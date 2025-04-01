package producer

import (
	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp"
)

type ProducerServer[T queue.PayloadIF] struct {
	ID     string              // producer id
	server *asynctcp.Transport // Prodcuer server instance
	q      *queue.Queue[T]     // thread safe message queue
}

// NewProducerServer creates a new ProducerServer to accept data of type T.
//
// It initializes a server to listen on the specified address and uses the given
// message queue for processing the data.
func NewProducerServer[T queue.PayloadIF](id string, q *queue.Queue[T]) (*ProducerServer[T], error) {
	server, err := asynctcp.NewTransport(id, asynctcp.TransportOpts{OnAcceptingConn: onAcceptingProdcuer, Ack: true})
	if err != nil {
		return nil, err
	}

	return &ProducerServer[T]{
		ID:     server.ID,
		server: server,
		q:      q,
	}, nil
}

// Listen starts the Pub server and begins populating data to the message queue.
func (t *ProducerServer[T]) Listen() error {
	err := t.server.Listen()
	t.startPopulatingQueue()
	return err
}

// Stop stops listening to new Producer connections
//
// Note: It still continues to serve existing connections
func (t *ProducerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

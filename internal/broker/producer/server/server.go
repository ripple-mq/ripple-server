package producer

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp"
)

type ProducerServer[T queue.PayloadIF] struct {
	ID        string              // producer id
	server    *asynctcp.Transport // Prodcuer server instance
	q         *queue.Queue[T]     // thread safe message queue
	topic     topic.TopicBucket
	ackServer *asynctcp.Transport
}

// NewProducerServer creates a new ProducerServer to accept data of type T.
//
// It initializes a server to listen on the specified address and uses the given
// message queue for processing the data.
func NewProducerServer[T queue.PayloadIF](id string, q *queue.Queue[T], topic topic.TopicBucket) (*ProducerServer[T], error) {
	server, _ := asynctcp.NewTransport(id, asynctcp.TransportOpts{OnAcceptingConn: onAcceptingProdcuer, Ack: true})
	ackServer, err := asynctcp.NewTransport(uuid.NewString())
	if err != nil {
		return nil, err
	}

	return &ProducerServer[T]{
		ID:        server.ID,
		server:    server,
		q:         q,
		topic:     topic,
		ackServer: ackServer,
	}, nil
}

// Listen starts the Pub server and begins populating data to the message queue.
func (t *ProducerServer[T]) Listen() error {
	if err := t.server.Listen(); err != nil {
		return fmt.Errorf("failed to start queue server: %v", err)
	}
	if err := t.ackServer.Listen(); err != nil {
		return fmt.Errorf("failed to start ack server: %v", err)
	}
	t.startPopulatingQueue()
	return nil
}

// Stop stops listening to new Producer connections
//
// Note: It still continues to serve existing connections
func (t *ProducerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

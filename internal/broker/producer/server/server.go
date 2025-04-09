package producer

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/ack"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/server/asynctcp"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type ProducerServer[T queue.PayloadIF] struct {
	ID         string              // producer id
	server     *asynctcp.Transport // Prodcuer server instance
	q          *queue.Queue[T]     // thread safe message queue
	topic      topic.TopicBucket
	ackHandler *ack.AcknowledgeHandler
	amILeader  *collection.ConcurrentValue[bool]
}

// NewProducerServer creates a new ProducerServer to accept data of type T.
//
// It initializes a server to listen on the specified address and uses the given
// message queue for processing the data.
func NewProducerServer[T queue.PayloadIF](id string, q *queue.Queue[T], topic topic.TopicBucket) (*ProducerServer[T], error) {
	server, err := asynctcp.NewTransport(id, asynctcp.TransportOpts{OnAcceptingConn: onAcceptingProdcuer, Ack: true})
	if err != nil {
		fmt.Println("error creating server, ", err)
	}
	ackServer, err := asynctcp.NewTransport(fmt.Sprintf("ack-%s", id), asynctcp.TransportOpts{OnAcceptingConn: onAcceptingProdcuer, Ack: true})
	if err != nil {
		fmt.Println("error creating ack server, ", err)
	}
	ackHandler := ack.NewAcknowledgeHandler(ackServer)

	return &ProducerServer[T]{
		ID:         server.ListenAddr.ID,
		server:     server,
		q:          q,
		topic:      topic,
		ackHandler: ackHandler,
		amILeader:  collection.NewConcurrentValue(false),
	}, nil
}

// Listen starts the Pub server and begins populating data to the message queue.
func (t *ProducerServer[T]) Listen() error {
	if err := t.server.Listen(); err != nil {
		return fmt.Errorf("failed to start queue server: %v", err)
	}
	if err := t.ackHandler.Run(); err != nil {
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

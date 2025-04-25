package server

import (
	"fmt"

	c "github.com/ripple-mq/ripple-server/internal/broker/consumer"
	cs "github.com/ripple-mq/ripple-server/internal/broker/consumer/server"
	p "github.com/ripple-mq/ripple-server/internal/broker/producer"
	ps "github.com/ripple-mq/ripple-server/internal/broker/producer/server"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/topic"
)

// Server holds bothe Pub/Sub server
type Server struct {
	PS *ps.ProducerServer[queue.Payload] // Producer server instance
	CS *cs.ConsumerServer[queue.Payload] // Consumer server instance
}

// NewServer creates a new Pub/Sub server instance with a producer and consumer.
//
// It initializes the producer and consumer but doesn't start listening.
// Call Listen() on the returned server to start listening and avoid a busy port error.
func NewServer(prodId, conId string, topic topic.TopicBucket) *Server {
	q := queue.NewQueue[queue.Payload]()
	p, _ := p.NewProducer(topic).ByteStreamingServer(prodId, q)
	c, _ := c.NewConsumer(topic).ByteStreamingServer(conId, q)
	return &Server{PS: p, CS: c}
}

// Listen starts the pub/sub servers on the specified ports.
func (t *Server) Listen() error {
	if err := t.PS.Listen(); err != nil {
		return fmt.Errorf("failed to start producer server: %v", err)
	}
	if err := t.CS.Listen(); err != nil {
		return fmt.Errorf("failed to start consumer server: %v", err)
	}
	return nil
}

// Stop stops existing pub/sub servers accepting new connections
//
// Note: existing connection will continue to serve only if p2p server is being used
func (t *Server) Stop() {
	t.PS.Stop()
	t.CS.Stop()
}

func (t *Server) InformLeaderStatus(val bool) {
	t.PS.InformLeaderStatus(val)
}

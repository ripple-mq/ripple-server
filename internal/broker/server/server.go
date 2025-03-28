package server

import (
	"fmt"

	c "github.com/ripple-mq/ripple-server/internal/broker/consumer"
	cs "github.com/ripple-mq/ripple-server/internal/broker/consumer/server"
	p "github.com/ripple-mq/ripple-server/internal/broker/producer"
	ps "github.com/ripple-mq/ripple-server/internal/broker/producer/server"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
)

// Server holds bothe Pub/Sub server
type Server struct {
	PS *ps.ProducerServer[queue.Payload] // Producer server instance
	CS *cs.ConsumerServer[queue.Payload] // Consumer server instance
}

// NewServer creates and returns a new Pub/Sub server instance.
//
// It initializes a new producer and consumer.
// Note: It doesn't start listening.
//
// It is advised to Listen() asap to avoid busy port error.
//
// Parameters:
//   - paddr (string): Producer server address.
//   - caddr (string): Consumer server address.
//
// Returns:
//   - *Server.
//
// Example usage:
//
//	server := NewServer("localhost:8080", "localhost:8081")
func NewServer(paddr, caddr string) *Server {
	q := queue.NewQueue[queue.Payload]()
	p, _ := p.NewProducer().ByteStreamingServer(paddr, q)
	c, _ := c.NewConsumer().ByteStreamingServer(caddr, q)
	return &Server{PS: p, CS: c}
}

// Listen spins up pub/sub servers at specified ports.
//
// Returns:
//   - error
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
// Note: existing connection will continue to serve
func (t *Server) Stop() {
	t.PS.Stop()
	t.CS.Stop()
}

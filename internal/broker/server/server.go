package server

import (
	"fmt"

	c "github.com/ripple-mq/ripple-server/internal/broker/consumer"
	cs "github.com/ripple-mq/ripple-server/internal/broker/consumer/server"
	p "github.com/ripple-mq/ripple-server/internal/broker/producer"
	ps "github.com/ripple-mq/ripple-server/internal/broker/producer/server"
)

type Server struct {
	PS *ps.ProducerServer[[]byte]
	CS *cs.ConsumerServer[[]byte]
}

func NewServer(paddr, caddr string) *Server {

	p, _ := p.NewProducer().ByteStreamingServer(paddr)
	c, _ := c.NewConsumer().ByteStreamingServer(caddr)
	return &Server{PS: p, CS: c}
}

func (t *Server) Listen() error {

	if err := t.PS.Listen(); err != nil {
		return fmt.Errorf("failed to start producer server: %v", err)
	}
	if err := t.CS.Listen(); err != nil {
		return fmt.Errorf("failed to start consumer server: %v", err)
	}
	return nil
}

func (t *Server) Stop() {
	t.PS.Stop()
	t.CS.Stop()
}

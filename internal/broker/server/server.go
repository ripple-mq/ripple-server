package server

import (
	"fmt"

	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	cs "github.com/ripple-mq/ripple-server/internal/broker/server/consumer"
	ps "github.com/ripple-mq/ripple-server/internal/broker/server/producer"
)

type Server struct {
	PS *ps.ProducerServer[[]byte]
	CS *cs.ConsumerServer[[]byte]
}

func NewServer(paddr, caddr string) *Server {
	q := queue.NewQueue[[]byte]()

	p, _ := ps.NewProducerServer(paddr, q)
	c, _ := cs.NewConsumerServer(caddr, q)
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

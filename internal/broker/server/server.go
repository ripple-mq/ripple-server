package server

import (
	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	cs "github.com/ripple-mq/ripple-server/internal/broker/server/consumer"
	ps "github.com/ripple-mq/ripple-server/internal/broker/server/producer"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (t *Server) Listen(paddr, caddr string) {
	q := queue.NewQueue[[]byte]()

	p, _ := ps.NewProducerServer(paddr, q)
	c, _ := cs.NewConsumerServer(caddr, q)

	if err := p.Listen(); err != nil {
		log.Errorf("failed to start producer server: %v", err)
	}
	if err := c.Listen(); err != nil {
		log.Errorf("failed to start consumer server: %v", err)
	}

}

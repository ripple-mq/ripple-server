package broker

import (
	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/broker/queue/server"
)

type Broker struct {
}

func NewBroker() *Broker {
	return &Broker{}
}

func (t *Broker) CreateAndRunQueue(paddr, caddr string) {
	q := queue.NewQueue[[]byte]()

	p, _ := server.NewProducerServer(paddr, q)
	c, _ := server.NewConsumerServer(caddr, q)

	if err := p.Listen(); err != nil {
		log.Errorf("failed to start producer server: %v", err)
	}
	if err := c.Listen(); err != nil {
		log.Errorf("failed to start consumer server: %v", err)
	}

}

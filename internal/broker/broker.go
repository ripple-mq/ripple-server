package broker

import (
	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	cs "github.com/ripple-mq/ripple-server/internal/broker/queue/server/consumer"
	ps "github.com/ripple-mq/ripple-server/internal/broker/queue/server/producer"
)

type Broker struct {
}

func NewBroker() *Broker {
	return &Broker{}
}

func (t *Broker) CreateAndRunQueue(paddr, caddr string) {
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

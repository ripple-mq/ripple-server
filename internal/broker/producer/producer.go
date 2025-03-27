package producer

import (
	"fmt"

	ps "github.com/ripple-mq/ripple-server/internal/broker/producer/server"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
)

type Producer struct {
}

func NewProducer() *Producer {
	return &Producer{}
}

func (t *Producer) ByteStreamingServer(addr string) (*ps.ProducerServer[[]byte], error) {
	server, err := ps.NewProducerServer(addr, queue.NewQueue[[]byte]())
	if err != nil {
		return nil, fmt.Errorf("failed to spin up server at %s: %v", addr, err)
	}
	return server, nil
}

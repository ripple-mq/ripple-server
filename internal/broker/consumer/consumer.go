package consumer

import (
	"fmt"

	cs "github.com/ripple-mq/ripple-server/internal/broker/consumer/server"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
)

type Consumer struct {
}

func NewConsumer() *Consumer {
	return &Consumer{}
}

func (t *Consumer) ByteStreamingServer(addr string) (*cs.ConsumerServer[[]byte], error) {
	server, err := cs.NewConsumerServer(addr, queue.NewQueue[[]byte]())
	if err != nil {
		return nil, fmt.Errorf("failed to spin up server at %s: %v", addr, err)
	}
	return server, nil
}

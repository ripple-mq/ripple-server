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

func (t *Producer) ByteStreamingServer(addr string, q *queue.Queue[queue.Payload]) (*ps.ProducerServer[queue.Payload], error) {
	server, err := ps.NewProducerServer(addr, q)
	if err != nil {
		return nil, fmt.Errorf("failed to spin up server at %s: %v", addr, err)
	}
	return server, nil
}

func (t *Producer) GetServerConnection(topic string, bucket string) (string, error) {
	return "", nil
}

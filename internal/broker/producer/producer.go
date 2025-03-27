package producer

import (
	"fmt"

	ps "github.com/ripple-mq/ripple-server/internal/broker/producer/server"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
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

func (t *Producer) GetServerConnection(topicName string, bucketName string) ([]byte, error) {
	topicPath := tp.TopicBucket{TopicName: topicName, BucketName: bucketName}.GetPath()
	lh := lighthouse.GetLightHouse()
	data, err := lh.ReadLeader(topicPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

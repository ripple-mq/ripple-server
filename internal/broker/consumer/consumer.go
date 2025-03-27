package consumer

import (
	"fmt"

	"github.com/ripple-mq/ripple-server/internal/broker/consumer/loadbalancer"
	cs "github.com/ripple-mq/ripple-server/internal/broker/consumer/server"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
)

type Consumer struct {
}

func NewConsumer() *Consumer {
	return &Consumer{}
}

func (t *Consumer) ByteStreamingServer(addr string, q *queue.Queue[queue.Payload]) (*cs.ConsumerServer[queue.Payload], error) {
	server, err := cs.NewConsumerServer(addr, q)
	if err != nil {
		return nil, fmt.Errorf("failed to spin up server at %s: %v", addr, err)
	}
	return server, nil
}

func (t *Consumer) GetServerConnection(topicName string, bucketName string) ([]byte, error) {
	topicPath := tp.TopicBucket{TopicName: topicName, BucketName: bucketName}.GetPath()
	lh := lighthouse.GetLightHouse()
	data, err := lh.ReadFollowers(topicPath)
	if err != nil {
		return nil, err
	}

	index := loadbalancer.NewReadReqLoadBalancer().GetIndex(len(data))
	return data[index], nil
}

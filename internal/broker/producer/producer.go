package producer

import (
	"fmt"

	ps "github.com/ripple-mq/ripple-server/internal/broker/producer/server"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	tp "github.com/ripple-mq/ripple-server/internal/topic"
)

type Producer struct {
	topic tp.TopicBucket
}

// NewProducer returns *Producer instance
func NewProducer(topic tp.TopicBucket) *Producer {
	return &Producer{topic: topic}
}

// ByteStreamingServer creates a Pub server that listens for messages and processes them using the provided message queue.
//
// It initializes a server to accept byte-streamed messages and uses the given queue to handle those messages.
// If server creation fails, an error is returned.
func (t *Producer) ByteStreamingServer(id string, q *queue.Queue[queue.Payload]) (*ps.ProducerServer[queue.Payload], error) {
	server, err := ps.NewProducerServer(id, q, t.topic)
	if err != nil {
		return nil, fmt.Errorf("failed to spin up server at %s: %v", id, err)
	}
	return server, nil
}

// GetServerConnection returns the write address for the given topic and bucket.
//
// It retrieves the leader server's address for the specified topic and bucket from Lighthouse.
func (t *Producer) GetServerConnection(topicName string, bucketName string) ([]byte, error) {
	topicPath := tp.TopicBucket{TopicName: topicName, BucketName: bucketName}.GetPath()
	lh := lighthouse.GetLightHouse()
	data, err := lh.ReadLeader(topicPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

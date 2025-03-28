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

// NewProducer returns *Producer instance
func NewProducer() *Producer {
	return &Producer{}
}

// ByteStreamingServer creates Pub server to accept messages
//
// Parameters:
//   - addr(string): address to listen at
//   - q(*queue.Queue[queue.Payload]): message queue of type `Payload`
//
// Returns:
//   - *ps.ProducerServer[queue.Payload]
//   - error
func (t *Producer) ByteStreamingServer(addr string, q *queue.Queue[queue.Payload]) (*ps.ProducerServer[queue.Payload], error) {
	server, err := ps.NewProducerServer(addr, q)
	if err != nil {
		return nil, fmt.Errorf("failed to spin up server at %s: %v", addr, err)
	}
	return server, nil
}

// GetServerConnection returns Write addr of given `topicName` , `'bucketName`
//
// Parameters:
//   - topicName(string)
//   - bucketName(string)
//
// Returns:
//   - []byte: PCServerAddr bytes
//   - error
func (t *Producer) GetServerConnection(topicName string, bucketName string) ([]byte, error) {
	topicPath := tp.TopicBucket{TopicName: topicName, BucketName: bucketName}.GetPath()
	lh := lighthouse.GetLightHouse()
	data, err := lh.ReadLeader(topicPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

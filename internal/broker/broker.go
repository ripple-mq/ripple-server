package broker

import (
	"fmt"

	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type Bucket[T any] struct {
	Topic string
	Name  string
	Queue *queue.Queue[T]
}

func (t Bucket[T]) Path() string {
	return fmt.Sprintf("/%s/%s", t.Topic, t.Name)
}

type Broker struct {
	Buckets collection.ConcurrentMap[string, *Bucket[any]]
}

var brokerInstance = newBroker()

func newBroker() *Broker {
	return &Broker{}
}

func GetBroker() *Broker {
	return brokerInstance
}

func (t *Broker) CreateBucket(bucket Bucket[any]) {

}

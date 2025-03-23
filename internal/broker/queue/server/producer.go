package queue

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type ProducerServer[T any] struct {
	listenAddr net.Addr
	server     *tcp.Transport
	q          *collection.ConcurrentList[T]
}

func NewProducerServer[T any](addr string, q *collection.ConcurrentList[T]) (*ProducerServer[T], error) {
	server, err := tcp.NewTransport(addr)
	if err != nil {
		return nil, err
	}

	return &ProducerServer[T]{
		listenAddr: server.ListenAddr,
		server:     server,
		q:          q,
	}, nil
}

func (t *ProducerServer[T]) Listen() error {
	return t.server.Listen()
}

func (t *ProducerServer[T]) Stop() {
	t.server.Stop()
}

func (t *ProducerServer[T]) PopulateQueueLoop() {
	go func() {
		for {
			var data T
			_, err := t.server.Consume(encoder.GOBDecoder{}, &data)
			if err != nil {
				log.Warnf("error reading data: %v", err)
			}
			t.q.Append(data)
		}
	}()
}

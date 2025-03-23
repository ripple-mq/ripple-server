package queue

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type AskQuery struct {
	Count int
}

type ConsumerServer[T any] struct {
	listenAddr net.Addr
	server     *tcp.Transport
	q          *collection.ConcurrentList[T]
}

func NewConsumerServer[T any](addr string, q *collection.ConcurrentList[T]) (*ConsumerServer[T], error) {
	server, err := tcp.NewTransport(addr)
	if err != nil {
		return nil, err
	}

	return &ConsumerServer[T]{
		listenAddr: server.ListenAddr,
		server:     server,
		q:          q,
	}, nil
}

func (t *ConsumerServer[T]) Listen() error {
	return t.server.Listen()
}

func (t *ConsumerServer[T]) Stop() {
	t.server.Stop()
}

func (t *ConsumerServer[T]) Stream() {
	go func() {
		for {
			var query AskQuery
			clientAddr, err := t.server.Consume(encoder.GOBDecoder{}, &query)
			if err != nil {
				log.Warnf("error reading data: %v", err)
			}
			t.StreamLoop(query, clientAddr)
		}
	}()
}

func (t *ConsumerServer[T]) StreamLoop(query AskQuery, clientAddr string) {
	//t.server.Send(clientAddr, )
}

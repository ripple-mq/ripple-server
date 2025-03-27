package producer

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
)

type ProducerServer[T any] struct {
	listenAddr net.Addr
	server     *tcp.Transport
	q          *queue.Queue[T]
}

func NewProducerServer[T any](addr string, q *queue.Queue[T]) (*ProducerServer[T], error) {
	server, err := tcp.NewTransport(addr, onAcceptingProdcuer)
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
	err := t.server.Listen()
	t.startPopulatingQueue()
	return err
}

func (t *ProducerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

package server

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
)

const ConsumerPath string = "/consumers"

// AskQuery needs a standard serialization to make it compatible with all language/frameworks
type AskQuery struct {
	Count int
	ID    string
}

type ConsumerServer[T any] struct {
	listenAddr net.Addr
	server     *tcp.Transport
	q          *queue.Queue[T]
}

func NewConsumerServer[T any](addr string, q *queue.Queue[T]) (*ConsumerServer[T], error) {
	server, err := tcp.NewTransport(addr, onAcceptingConsumer)
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
	err := t.server.Listen()
	t.startAcceptingConsumeReq()
	return err
}

func (t *ConsumerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

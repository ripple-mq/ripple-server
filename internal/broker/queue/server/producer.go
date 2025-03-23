package server

import (
	"bytes"
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
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
	t.server.Stop()
}

func (t *ProducerServer[T]) startPopulatingQueue() {
	go func() {
		for {
			var data T
			_, err := t.server.Consume(encoder.GOBDecoder{}, &data)
			if err != nil {
				log.Warnf("error reading data: %v", err)
			}
			t.q.Push(data)
		}
	}()
}

func onAcceptingProdcuer(conn net.Conn, msg []byte) {
	var MSG string
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(msg), &MSG)
	if err != nil {
		return
	}
	log.Infof("Accepting producer: %v, message: %s", conn, MSG)
}

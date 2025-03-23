package server

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
)

const ConsumerPath string = "/consumers"

// AskQuery needs a standard serialization to make it compatible with all language/frameworks
type AskQuery struct {
	Count int
	Id    string
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
	t.startStreaming()
	return err
}

func (t *ConsumerServer[T]) Stop() {
	t.server.Stop()
}

func (t *ConsumerServer[T]) startStreaming() {
	go func() {
		for {
			var query AskQuery
			clientAddr, err := t.server.Consume(encoder.GOBDecoder{}, &query)
			if err != nil {
				continue
			}
			go t.StreamingLoop(query, clientAddr)
		}
	}()
}

func (t *ConsumerServer[T]) StreamingLoop(query AskQuery, clientAddr string) {
	lh, _ := lighthouse.GetLightHouse()
	data, _ := lh.Read(lighthouse.Path{Base: fmt.Sprintf("%s/%s", ConsumerPath, query.Id)})
	offset, _ := strconv.Atoi(string(data))
	defer lh.UpdateZnode(getConsumerPath(query.Id), strconv.Itoa(offset+query.Count))

	for {
		messages := t.q.SubArray(offset, offset+query.Count)
		if len(messages) == 0 {
			continue
		}
		t.server.Send(clientAddr, struct{}{}, messages)
		offset += len(messages)
	}
}

func onAcceptingConsumer(conn net.Conn, id []byte) {
	var ID string
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(id), &ID)
	if err != nil {
		log.Warnf("error reading data: %v", err)
	}
	log.Infof("Accepting consumer: %s %s", conn.RemoteAddr(), ID)
	registerConsumer(ID)
}

func registerConsumer(id string) {
	lh, _ := lighthouse.GetLightHouse()
	lh.RegisterSequential(getConsumerPath(id), strconv.Itoa(0))
}

func getConsumerPath(id string) lighthouse.Path {
	return lighthouse.Path{Base: fmt.Sprintf("%s/%s", ConsumerPath, id)}
}

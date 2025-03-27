package consumer

import (
	"bytes"
	"net"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
)

// AskQuery needs a standard serialization to make it compatible with all language/frameworks
type AskQuery struct {
	Count int
	ID    string
}

const ConsumerPath string = "/consumers"

type Consumer struct {
}

func (t *ConsumerServer[T]) startAcceptingConsumeReq() {
	go func() {
		for {
			var query AskQuery
			clientAddr, err := t.server.Consume(encoder.GOBDecoder{}, &query)
			if err != nil {
				continue
			}
			go t.handleConsumeReq(query, clientAddr)
		}
	}()
}

// TODO: offset will go wrong once i introduce TTL
func (t *ConsumerServer[T]) handleConsumeReq(query AskQuery, clientAddr string) {
	lh := lighthouse.GetLightHouse()
	data, _ := lh.Read(Consumer{}.getConsumerPath(query.ID))
	offset, _ := strconv.Atoi(string(data))
	defer lh.Write(Consumer{}.getConsumerPath(query.ID), strconv.Itoa(offset+query.Count))

	for {
		messages := t.q.SubArray(offset, offset+query.Count)
		if len(messages) == 0 {
			continue
		}
		if err := t.server.Send(clientAddr, struct{}{}, messages); err != nil {
			log.Errorf("failed to send: %v", err)
		}
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
	Consumer{}.registerConsumer(ID)
}

func (t Consumer) registerConsumer(id string) {
	lh := lighthouse.GetLightHouse()
	lh.RegisterSequential(t.getConsumerPath(id), strconv.Itoa(0))
}

func (t Consumer) getConsumerPath(id string) utils.Path {
	return utils.PathBuilder{}.Base(utils.Root()).CD(ConsumerPath).CD(id).Create()
}

package server

import (
	"bytes"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/comm"
)

// AskQuery needs a standard serialization to make it compatible with all language/frameworks
type AskQuery struct {
	Count int
	ID    string
}

const ConsumerPath string = "consumers"

// startAcceptingConsumeReq listens for incoming Ask queries and handles each request asynchronously.
//
// It continuously consumes requests and processes asynchronously. Any failed request
// is ignored, and the function keeps listening for new ones.
func (t *ConsumerServer[T]) startAcceptingConsumeReq() {
	go func() {
		for {
			var query AskQuery
			clientAddr, err := t.server.Consume(encoder.GOBDecoder{}, &query)
			if err != nil {
				continue
			}
			go t.handleConsumeReq(query, clientAddr.Addr)
		}
	}()
}

// handleConsumeReq streams messages to the client in the specified batch size.
//
// It retrieves a batch of messages based on the current offset and sends them to the client.
// The offset is updated after each batch.
//
// TODO: Offset handling may need adjustment once TTL is introduced.
func (t *ConsumerServer[T]) handleConsumeReq(query AskQuery, clientAddr string) {
	lh := lighthouse.GetLightHouse()
	data, _ := lh.Read(getConsumerPath(query.ID))
	offset, _ := strconv.Atoi(string(data))

	for {
		messages := t.q.SubArray(offset, offset+query.Count)
		// Should i terminate current response if no message available for now
		if len(messages) == 0 {
			continue
		}
		if err := t.server.Send(clientAddr, struct{}{}, messages); err != nil {
			log.Errorf("failed to send: %v", err)
			break
		}
		offset += len(messages)
		go lh.Write(getConsumerPath(query.ID), strconv.Itoa(offset+query.Count))
	}
}

// onAcceptingConsumer handles a new consumer connection and registers the consumer.
func onAcceptingConsumer(msg comm.Message) {
	var ID string
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(msg.Payload), &ID)
	if err != nil {
		log.Warnf("error reading data: %v", err)
	}
	log.Infof("Accepting consumer: %s %s", msg.RemoteAddr, ID)
	registerConsumer(ID)
}

func registerConsumer(id string) {
	lh := lighthouse.GetLightHouse()
	path := getConsumerPath(id)
	lh.RegisterSequential(path, strconv.Itoa(0))
}

func getConsumerPath(id string) utils.Path {
	return utils.PathBuilder{}.Base(utils.Root()).CD(ConsumerPath).CD(id).Create()
}

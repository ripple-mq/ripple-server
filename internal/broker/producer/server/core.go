package producer

import (
	"bytes"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/ack"
	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/internal/processor"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	acomm "github.com/ripple-mq/ripple-server/pkg/server/asynctcp/comm"
)

// startPopulatingQueue consumes messages from the server and pushes them to the queue asynchronously.
//
// It continuously listens for incoming messages, decodes them, and adds them to the queue.
// Triggers gossip asynchronously to spread data to followers & handle acknowledgment
func (t *ProducerServer[T]) startPopulatingQueue() {
	go func() {
		for {
			var data T
			clientAddr, err := t.server.Consume(encoder.GOBDecoder{}, &data)
			if err != nil {
				log.Warnf("error reading data: %v", err)
			}

			t.q.Push(data)
			go t.GossipPush(clientAddr, data)
		}
	}()
}

// onAcceptingProducer handles new Producer connections and logs the connection details along with the metadata message.
//
// Note: it will be executed for every new producer connection
func onAcceptingProdcuer(msg acomm.Message) {
	var MSG string
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(msg.Payload), &MSG)
	if err != nil {
		return
	}
	log.Infof("Accepting producer: %s  message: %s", msg.RemoteAddr, MSG)
}

func (t *ProducerServer[T]) InformLeaderStatus() {
	t.amILeader.Set(true)
}

// GossipPush handles pushing messages to replicas & acknowledgement.
// Leader pushes data to replicas & adds ack task to queue.
// Followers acknowledges client (leader).
func (t *ProducerServer[T]) GossipPush(clientAddr acomm.ServerAddr, data queue.PayloadIF) {
	if !t.amILeader.Get() {
		t.ackHandler.P2PServer.Send(clientAddr.Addr, struct{}{}, queue.Ack{Id: data.GetID()})
		return
	}

	// get all followers for t.topicBucket
	lh := lighthouse.GetLightHouse()
	followers, err := lh.ReadFollowers(t.topic.GetPath())
	if err != nil {
		return
	}

	// broadcasting messages to all followers
	var followersAddr []comm.PCServerID
	for _, s := range followers {
		addr, err := comm.DecodeToPCServerID(s)
		if err != nil || (addr.BrokerAddr == t.server.ListenAddr.Addr && addr.ProducerID == t.ID) {
			continue
		}
		followersAddr = append(followersAddr, addr)
	}

	task := ack.Task{
		AckHandler:    t.ackHandler,
		AckClientAddr: clientAddr.Addr,
		Receivers:     followersAddr,
		Data:          data,
		Wg:            &sync.WaitGroup{},
		MsgID:         data.GetID(),
	}

	processor.GetProcessor().Add(task)
}

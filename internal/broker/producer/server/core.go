package producer

import (
	"bytes"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/internal/processor"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	acomm "github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp/comm"
)

// startPopulatingQueue consumes messages from the server and pushes them to the queue asynchronously.
//
// It continuously listens for incoming messages, decodes them, and adds them to the queue.
// The function runs in a separate goroutine to avoid blocking other operations.
func (t *ProducerServer[T]) startPopulatingQueue() {
	go func() {
		for {
			var data T
			clientAddr, err := t.server.Consume(encoder.GOBDecoder{}, &data)
			if err != nil {
				log.Warnf("error reading data: %v", err)
			}

			t.q.Push(data)

			// get all followers for t.topicBucket
			lh := lighthouse.GetLightHouse()
			followers, err := lh.ReadFollowers(t.topic.GetPath())
			if err != nil {
				continue
			}

			// broadcasting messages to all followers
			var followersAddr []comm.PCServerID
			for _, s := range followers {
				addr, err := comm.DecodeToPCServerID(s)
				if err != nil {
					continue
				}
				followersAddr = append(followersAddr, addr)
			}

			task := Task{
				Server:        t.ackServer,
				AckClientAddr: clientAddr,
				Receivers:     followersAddr,
				Data:          data,
				Wg:            &sync.WaitGroup{},
				MsgID:         data.GetID(),
			}

			processor.GetProcessor().Add(task)

			// followers
			t.server.SendToAsync(clientAddr, data, struct{}{}, queue.Ack{Id: data.GetID()})
		}
	}()
}

// onAcceptingProducer handles new Producer connections and logs the connection details along with the metadata message.
//
// Note: it will be executed for every new connection
func onAcceptingProdcuer(msg acomm.Message) {
	var MSG string
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(msg.Payload), &MSG)
	if err != nil {
		return
	}
	log.Infof("Accepting producer: %s  message: %s", msg.RemoteAddr, MSG)
}

package broker_test

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/internal/topic"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
)

type AskQuery struct {
	Count int
	ID    string
}

func TestCreateAndRunQueue(t *testing.T) {
	lighthouse.GetLightHouse()
	b := broker.NewBroker(topic.TopicBucket{TopicName: "t0", BucketName: "b0"})

	pId, cId := "p1", "c1"
	b.Run(pId, cId)

	prod, _ := tcp.NewTransport(":9000", func(conn net.Conn, msg []byte) {})
	cons, _ := tcp.NewTransport(":9001", func(conn net.Conn, msg []byte) {})

	go func() {
		for {
			var msg queue.Ack
			_, err := prod.Consume(encoder.GOBDecoder{}, &msg)
			if err != nil {
				log.Warnf("ack error : %v", err)
			}
			fmt.Println("Producer: ack = ", msg.Id)
		}
	}()

	go func() {
		i := 0
		for {
			var buff bytes.Buffer
			err := encoder.GOBEncoder{}.Encode((fmt.Sprintf("message: %d", i)), &buff)
			if err != nil {
				t.Errorf("failed to encode: %v", err)
			}
			if err := prod.SendToAsync(pId, struct{}{}, queue.Payload{Id: int32(i), Data: buff.Bytes()}); err != nil {
				t.Errorf("failed to send: %v", err)
			}
			i++
			if i == 20 {
				time.Sleep(2 * time.Second)
			}
		}
	}()

	if err := cons.SendToAsync(cId, "0", AskQuery{Count: 4, ID: strconv.Itoa(0)}); err != nil {
		t.Errorf("failed to send: %v", err)
	}

	go func() {
		for {
			var msg []queue.Payload
			addr, err := cons.Consume(encoder.GOBDecoder{}, &msg)
			if err != nil {
				log.Warnf("error: %v", err)
			}
			var msgs []string
			for _, i := range msg {
				var m string
				err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(i.Data), &m)
				if err != nil {
					t.Errorf("failed to decode: %v", err)
				}
				msgs = append(msgs, m)
			}
			log.Infof("Consumer: %s %s ", addr, msgs)
		}
	}()

	time.Sleep(5 * time.Second)
}

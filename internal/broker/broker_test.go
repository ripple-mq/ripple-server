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
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
	"github.com/ripple-mq/ripple-server/pkg/utils"
)

type AskQuery struct {
	Count int
	ID    string
}

func TestCreateAndRunQueue(t *testing.T) {

	lh, _ := lighthouse.GetLightHouse()
	lh.Connect()

	b := broker.NewBroker()
	paddr, caddr := utils.RandLocalAddr(), utils.RandLocalAddr()
	b.CreateAndRunQueue(paddr, caddr)

	client, _ := tcp.NewTransport(":9000", func(conn net.Conn, msg []byte) {})

	go func() {
		i := 0
		for {
			var buff bytes.Buffer
			err := encoder.GOBEncoder{}.Encode((fmt.Sprintf("message: %d", i)), &buff)
			if err != nil {
				t.Errorf("failed to encode: %v", err)
			}
			if err := client.Send(paddr, struct{}{}, buff.Bytes()); err != nil {
				t.Errorf("failed to send: %v", err)
			}
			i++
		}
	}()

	if err := client.Send(caddr, "0", AskQuery{Count: 4, ID: strconv.Itoa(0)}); err != nil {
		t.Errorf("failed to send: %v", err)
	}

	go func() {
		for {
			var msg [][]byte
			addr, err := client.Consume(encoder.GOBDecoder{}, &msg)
			if err != nil {
				log.Warnf("error: %v", err)
			}
			var msgs []string
			for _, i := range msg {
				var m string
				err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(i), &m)
				if err != nil {
					t.Errorf("failed to decode: %v", err)
				}
				msgs = append(msgs, m)
			}
			log.Infof("Consumer: %s %s ", addr, msgs)
		}
	}()

	time.Sleep(3 * time.Second)
}

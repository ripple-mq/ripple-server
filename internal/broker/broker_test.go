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
	Id    string
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
			encoder.GOBEncoder{}.Encode((fmt.Sprintf("message: %d", i)), &buff)
			client.Send(paddr, struct{}{}, buff.Bytes())
			i++
		}
	}()

	client.Send(caddr, "0", AskQuery{Count: 4, Id: strconv.Itoa(0)})

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
				encoder.GOBDecoder{}.Decode(bytes.NewBuffer(i), &m)
				msgs = append(msgs, m)
			}
			log.Infof("Consumer: %s %s ", addr, msgs)
		}
	}()

	time.Sleep(3 * time.Second)
}

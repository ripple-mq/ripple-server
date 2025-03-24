package server

import (
	"bytes"
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
)

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

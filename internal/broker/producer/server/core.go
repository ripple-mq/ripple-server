package producer

import (
	"bytes"
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
)

// startPopulatingQueue consumes messages from the server and pushes them to the queue asynchronously.
//
// It continuously listens for incoming messages, decodes them, and adds them to the queue.
// The function runs in a separate goroutine to avoid blocking other operations.
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

// onAcceptingProducer handles new Producer connections and logs the connection details along with the metadata message.
//
// Note: it will be executed for every new connection
func onAcceptingProdcuer(conn net.Conn, msg []byte) {
	var MSG string
	err := encoder.GOBDecoder{}.Decode(bytes.NewBuffer(msg), &MSG)
	if err != nil {
		return
	}
	log.Infof("Accepting producer: %v, message: %s", conn, MSG)
}

package gossip

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp"
)

type Task struct {
	Server        *asynctcp.Transport
	AckClientAddr string
	ReceiverAddr  string
	ReceiverID    string
	Data          any
}

func (t Task) Exec() {
	t.Server.SendToAsync(t.ReceiverAddr, t.ReceiverID, struct{}{}, t.Data)
	go func() {
		timeout := time.After(3 * time.Second)
		var ack queue.Ack
		_, err := t.Server.Consume(encoder.GOBDecoder{}, &ack, timeout)
		if err != nil {
			log.Errorf("failed to receive acknowledgement: %v", err)
			return
		}
		t.Server.Send(t.AckClientAddr, struct{}{}, ack.Id)
	}()
}

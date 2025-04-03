package producer

import (
	"context"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type Task struct {
	Server        *asynctcp.Transport
	AckClientAddr string
	Receivers     []comm.PCServerID
	Data          any
	Wg            *sync.WaitGroup
	MsgID         int32
}

func (t Task) Exec() error {
	for _, addr := range t.Receivers {
		if err := t.Server.SendToAsync(addr.BrokerAddr, addr.ProducerID, struct{}{}, t.Data); err != nil {
			log.Errorf("failed to send data to follower: %v", addr)
			continue
		}
		t.Wg.Add(1)
	}
	ctx, cancel := GetAknowledge().Add(t.MsgID, t.Wg)
	GetAknowledge().AcknowledgeClient(ctx, cancel, t.MsgID, func() {
		t.Server.Send(t.AckClientAddr, struct{}{}, t.Data)
	})
	return nil
}

type Counter struct {
	WG   *sync.WaitGroup
	Quit chan struct{}
}

type Acknowledge struct {
	server      *asynctcp.Transport
	ackMessages *collection.ConcurrentMap[int32, *Counter]
}

var ackInstance *Acknowledge

func GetAknowledge() *Acknowledge {
	if ackInstance != nil {
		return ackInstance
	}
	ackInstance = newAcknowledge()
	return ackInstance
}

func newAcknowledge() *Acknowledge {
	server, _ := asynctcp.NewTransport(uuid.NewString())
	return &Acknowledge{server: server}
}

func (t *Acknowledge) Run() {
	go func() {
		for {
			var ack queue.Ack
			_, err := t.server.Consume(encoder.GOBDecoder{}, &ack)
			if err != nil {
				log.Errorf("failed to receive acknowledgement: %v", err)
			}
			msgKey, err := t.ackMessages.Get(ack.Id)
			if err != nil {
				continue
			}
			msgKey.WG.Done()
		}

	}()
}

func (t *Acknowledge) Add(msgKey int32, wg *sync.WaitGroup) (context.Context, context.CancelFunc) {
	counter := Counter{WG: wg, Quit: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.ackMessages.Set(msgKey, &counter)
	return ctx, cancel
}

func (t *Acknowledge) AcknowledgeClient(ctx context.Context, cancel context.CancelFunc, msgKey int32, ack func()) {
	counter, err := t.ackMessages.Get(msgKey)
	if err != nil {
		log.Errorf("Message not added to Acknowledgement handler yet, %v", err)
	}
	go func() {
		counter.WG.Wait()
		close(counter.Quit)
	}()

	go func() {
		select {
		case <-counter.Quit:
			ack()
		case <-ctx.Done():
			log.Errorf("failed to recieve ack for %d, timeout", msgKey)
			cancel()
		}
	}()
}

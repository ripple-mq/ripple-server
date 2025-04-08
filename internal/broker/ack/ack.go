package ack

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/asynctcp"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type Task struct {
	AckHandler    *AcknowledgeHandler
	AckClientAddr string
	Receivers     []comm.PCServerID
	Data          queue.PayloadIF
	Wg            *sync.WaitGroup
	MsgID         int32
}

func (t Task) Exec() error {
	count := 0
	for _, addr := range t.Receivers {
		fmt.Println("follower: ", addr)
		if err := t.AckHandler.P2PServer.SendToAsync(addr.BrokerAddr, addr.ProducerID, struct{}{}, t.Data); err != nil {
			log.Errorf("failed to send data to follower: %v", addr)
			continue
		}
		t.Wg.Add(1)
		count++
	}
	if count == 0 {
		return nil
	}
	ctx, cancel := t.AckHandler.Add(t.MsgID, t.Wg)
	t.AckHandler.AcknowledgeClient(ctx, cancel, t.MsgID, func() {
		t.AckHandler.P2PServer.Send(t.AckClientAddr, struct{}{}, queue.Ack{Id: t.Data.GetID()})
	})
	return nil
}

type Counter struct {
	WG   *sync.WaitGroup
	Quit chan struct{}
}

type AcknowledgeHandler struct {
	ackMessages *collection.ConcurrentMap[int32, *Counter]
	P2PServer   *asynctcp.Transport
}

func NewAcknowledgeHandler(p2pServer *asynctcp.Transport) *AcknowledgeHandler {
	return &AcknowledgeHandler{ackMessages: collection.NewConcurrentMap[int32, *Counter](), P2PServer: p2pServer}
}

func (t *AcknowledgeHandler) Run() error {
	go func() {
		for {
			var ack queue.Ack
			_, err := t.P2PServer.Consume(encoder.GOBDecoder{}, &ack)
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
	return nil
}

func (t *AcknowledgeHandler) Add(msgKey int32, wg *sync.WaitGroup) (context.Context, context.CancelFunc) {
	counter := Counter{WG: wg, Quit: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.ackMessages.Set(msgKey, &counter)
	return ctx, cancel
}

func (t *AcknowledgeHandler) AcknowledgeClient(ctx context.Context, cancel context.CancelFunc, msgKey int32, ack func()) {
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

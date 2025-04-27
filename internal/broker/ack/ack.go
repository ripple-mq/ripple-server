package ack

/*
	NOTES:
		- Send the data from p2p client as asyncServer is not able to initalize new connection
		- Set ShouldClientHandleConn true to listen for ack from follower
*/

import (
	"context"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/comm"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/encoder"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
	"github.com/ripple-mq/ripple-server/pkg/server/asynctcp"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

// Task represents a unit of work that sends data to multiple followers and handles acknowledgments.
type Task struct {
	AckHandler    *AcknowledgeHandler // AcknowledgeHandler for managing acknowledgment
	AckClientAddr string              // Client (producer) address for final acknowledgment
	Receivers     []comm.PCServerID   // List of followers (consumers) to send data to
	Data          queue.PayloadIF     // The data to be forwarded to followers
	Wg            *sync.WaitGroup     // WaitGroup to wait for all acknowledgment responses
	MsgID         int32               // Unique message identifier for this task
}

// Exec sends data to all followers asynchronously and manages acknowledgment.
// Once all followers acknowledge, it sends the final acknowledgment to the client.
//
// Note: final ack to client is disabled for now
func (t Task) Exec() error {
	count := 0
	for _, addr := range t.Receivers {
		if err := t.AckHandler.P2PClient.SendToAsync(addr.BrokerAddr, addr.ProducerID, struct{}{}, t.Data); err != nil {
			log.Errorf("failed to send data to follower: %v", addr)
			continue
		}
		t.Wg.Add(1)
		count++
	}

	return nil
}

type Counter struct {
	WG   *sync.WaitGroup
	Quit chan struct{}
}

type AcknowledgeHandler struct {
	ackMessages *collection.ConcurrentMap[int32, *Counter]
	P2PClient   *tcp.Transport
	AsyncServer *asynctcp.Transport
}

func NewAcknowledgeHandler(p2pClient *tcp.Transport, asyncServer *asynctcp.Transport) *AcknowledgeHandler {
	return &AcknowledgeHandler{ackMessages: collection.NewConcurrentMap[int32, *Counter](), P2PClient: p2pClient, AsyncServer: asyncServer}
}

// Run starts a goroutine that listens for acknowledgments from the P2P server.
// Upon receiving an acknowledgment, it marks the corresponding task as done.
func (t *AcknowledgeHandler) Run() error {
	go func() {
		for {
			var ack queue.Ack
			_, err := t.P2PClient.Consume(encoder.GOBDecoder{}, &ack)
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

// Add stores a new acknowledgment task with a timeout context and wait group.
// It associates the given message key with a counter that tracks acknowledgment progress.
func (t *AcknowledgeHandler) Add(msgKey int32, wg *sync.WaitGroup) (context.Context, context.CancelFunc) {
	counter := Counter{WG: wg, Quit: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.ackMessages.Set(msgKey, &counter)
	return ctx, cancel
}

// AcknowledgeClient waits for the acknowledgment of a message, executing the provided callback
// once the acknowledgment is received, or canceling if the timeout context is reached.
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

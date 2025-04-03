package comm

import (
	"fmt"
	"time"

	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type ServerAddr struct {
	Addr string
	ID   string
}

type Message struct {
	RemoteAddr string
	Payload    []byte
}

type Subscriber struct {
	ID                string
	IncommingMsgQueue chan Message
	OnAcceptingConn   func(Message)
	GreetStatus       *collection.ConcurrentValue[bool]
}

func NewSubscriber(id string, OnAcceptingConn func(Message)) *Subscriber {
	return &Subscriber{
		ID:                id,
		IncommingMsgQueue: make(chan Message, 10),
		OnAcceptingConn:   OnAcceptingConn,
		GreetStatus:       collection.NewConcurrentValue(false),
	}
}

func (t *Subscriber) Push(msg Message) {
	t.IncommingMsgQueue <- msg
}

func (t *Subscriber) Poll(timeout ...<-chan time.Time) (Message, error) {
	if len(timeout) == 0 {
		return <-t.IncommingMsgQueue, nil
	}
	select {
	case <-timeout[0]:
		var null Message
		return null, fmt.Errorf("failed to poll data, timeout")
	case value := <-t.IncommingMsgQueue:
		return value, nil
	}

}

func (t *Subscriber) Greet(msg Message) {
	t.OnAcceptingConn(msg)
}

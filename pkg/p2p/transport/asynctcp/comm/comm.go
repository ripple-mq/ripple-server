package comm

import (
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

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

func (t *Subscriber) Poll() Message {
	return <-t.IncommingMsgQueue
}

func (t *Subscriber) Greet(msg Message) {
	t.OnAcceptingConn(msg)
}

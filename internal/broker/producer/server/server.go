package producer

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/broker/queue"
	"github.com/ripple-mq/ripple-server/pkg/p2p/transport/tcp"
)

type ProducerServer[T queue.PayloadIF] struct {
	listenAddr net.Addr        // listening address
	server     *tcp.Transport  // Prodcuer server instance
	q          *queue.Queue[T] // thread safe message queue
}

// NewProducerServer creates a new ProducerServer to accept data of type T.
//
// It initializes a server to listen on the specified address and uses the given
// message queue for processing the data.
func NewProducerServer[T queue.PayloadIF](addr string, q *queue.Queue[T]) (*ProducerServer[T], error) {
	server, err := tcp.NewTransport(addr, onAcceptingProdcuer, tcp.TransportOpts{ShouldClientHandleConn: true, Ack: true})
	if err != nil {
		return nil, err
	}

	return &ProducerServer[T]{
		listenAddr: server.ListenAddr,
		server:     server,
		q:          q,
	}, nil
}

// Listen starts the Pub server and begins populating data to the message queue.
func (t *ProducerServer[T]) Listen() error {
	err := t.server.Listen()
	t.startPopulatingQueue()
	return err
}

// Stop stops listening to new Producer connections
//
// Note: It still continues to serve existing connections
func (t *ProducerServer[T]) Stop() {
	if err := t.server.Stop(); err != nil {
		log.Errorf("failed to stop: %v", err)
	}
}

package tcp

import (
	"errors"
	"net"

	"github.com/charmbracelet/log"
)

// connectionLoop continuously accepts incoming connections on the listener,
// adds them to the connection pool, and handles each connection in a new goroutine.
func (t *Transport) connectionLoop() {
	defer func() {
		log.Infof("Shutting down server: %s", t.ListenAddr)
		t.wg.Done()
	}()

	t.wg.Add(1)
	for {

		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.Warnf("failed to establish connection with %s", t.ListenAddr.String())
		}

		t.addConnection(conn)
		go t.handleConnection(conn)

	}
}

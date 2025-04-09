package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/peer"
)

// handleConnection processes incoming data from a network connection, handling message length and payload.
// It calls the OnAcceptingConn function on the first message and enqueues subsequent messages to the incoming queue.
func (t *Transport) handleConnection(conn net.Conn) {
	isOnConnExecuted := false
	defer func() {
		err := t.dropConnection(conn.RemoteAddr().String())
		if err != nil {
			log.Debugf("Error while dropping connection: %v", err)
		}
		log.Debugf("Dropping connection: %s\n", conn.RemoteAddr())
	}()

	for {
		// first 4 bytes are used for message length
		lengthBytes := make([]byte, 4)
		_, err := conn.Read(lengthBytes)
		length := binary.BigEndian.Uint32(lengthBytes)

		if err == io.EOF {
			return
		}

		var buffer = make([]byte, length)
		n, _ := conn.Read(buffer)

		if !isOnConnExecuted {
			t.OnAcceptingConn(conn, buffer[:n])
			isOnConnExecuted = true
			continue
		}

		log.Debugf("Recieved message of length %d from %s", n, conn.RemoteAddr().String())
		t.IncommingMsgQueue <- Message{RemoteAddr: conn.RemoteAddr().String(), Payload: buffer[:n]}

	}
}

// getConnection retrieves an existing connection for the given address from the PeersMap.
// Returns the peer associated with the address or an error if no connection exists.
func (t *Transport) getConnection(addr string) (peer.Peer, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tcpAddr := t.buildAddress(addr)
	if conn := t.PeersMap[tcpAddr]; conn != nil {
		return conn, nil
	}
	return nil, fmt.Errorf("failed to get connection")
}

// addConnection adds a new connection to the PeersMap and returns the corresponding peer.
func (t *Transport) addConnection(conn net.Conn) peer.Peer {
	t.mu.Lock()
	x := conn.RemoteAddr().String()
	t.PeersMap[x] = &peer.TCPPeer{Conn: conn}
	res := t.PeersMap[x]
	t.mu.Unlock()
	return res
}

// dropConnection removes a connection for the given address from the PeersMap and closes it.
// Returns an error if no connection exists for the address.
func (t *Transport) dropConnection(addr string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	ResolveTCPAddr := t.buildAddress(addr)
	if conn := t.PeersMap[ResolveTCPAddr]; conn != nil {
		delete(t.PeersMap, conn.GetAddress().String())
		return conn.Close()
	}
	return fmt.Errorf("no existing connections found")
}

func (t *Transport) buildAddress(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr

}

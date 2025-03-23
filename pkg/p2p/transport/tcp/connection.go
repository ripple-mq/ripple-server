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

func (t *Transport) handleConnection(conn net.Conn) {
	t.OnAcceptingConn(conn)
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

		log.Debugf("Recieved message of length %d from %s", n, conn.RemoteAddr().String())
		t.IncommingMsgQueue <- Message{RemoteAddr: conn.RemoteAddr().String(), Payload: buffer[:n]}

	}
}

func (t *Transport) getConnection(addr string) (peer.Peer, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tcpAddr := t.buildAddress(addr)
	if conn := t.PeersMap[tcpAddr]; conn != nil {
		return conn, nil
	}
	return nil, fmt.Errorf("failed to get connection")
}

func (t *Transport) addConnection(conn net.Conn) peer.Peer {
	t.mu.Lock()
	t.PeersMap[conn.RemoteAddr().String()] = &peer.TCPPeer{Conn: conn}
	res := t.PeersMap[conn.RemoteAddr().String()]
	t.mu.Unlock()
	return res
}

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

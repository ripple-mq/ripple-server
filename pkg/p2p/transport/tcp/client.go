package tcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/peer"
)

func (t *Transport) send(addr string, data any) error {
	peerNode, err := t.dial(addr)
	if err != nil {
		return err
	}
	var msg bytes.Buffer
	_ = t.Encoder.Encode(data, &msg)

	length := uint32(len(msg.Bytes()))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	_, err = peerNode.GetConnection().Write(lengthBytes)
	if err != nil {
		return fmt.Errorf("failed to send length %s, %v", addr, err)
	}
	n, err := peerNode.GetConnection().Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send data to %s, %v", addr, err)
	}
	log.Debugf("successfully wrote %d bytes to %s", n, addr)

	return nil
}

func (t *Transport) dial(addr string) (peer.Peer, error) {
	peerNode, _ := t.getConnection(addr)
	if peerNode != nil {
		return peerNode, nil
	}
	conn, err := net.Dial(t.ListenAddr.Network(), addr)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection with %s", addr)
	}
	t.addConnection(conn)
	go t.handleConnection(conn)

	return t.addConnection(conn), nil
}

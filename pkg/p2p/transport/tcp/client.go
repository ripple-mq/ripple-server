package tcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/p2p/peer"
)

// send establishes a connection to the given address, then encodes and sends the data to the peer.
func (t *Transport) send(addr string, metadata any, data any) error {
	peerNode, err := t.dial(addr, metadata)
	if err != nil {
		return err
	}

	t.write(peerNode, data)
	return nil
}

// write encodes the given data and sends it to the specified peer node,
// first sending the length of the message followed by the message itself.
func (t *Transport) write(peerNode peer.Peer, data any) {
	var msg bytes.Buffer
	_ = t.Encoder.Encode(data, &msg)

	length := uint32(len(msg.Bytes()))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	_, err := peerNode.GetConnection().Write(lengthBytes)
	if err != nil {
		log.Errorf("failed to send: %v", err)
	}
	_, err = peerNode.GetConnection().Write(msg.Bytes())
	if err != nil {
		log.Errorf("failed to send data: %v", err)
	}
}

// dial attempts to establish a connection to the given address and returns the associated peer node.
//   - If a connection already exists, it returns the existing peer; otherwise, it establishes a new one.
//   - Based on config, it can start listening on same connection.
func (t *Transport) dial(addr string, metadata any) (peer.Peer, error) {
	peerNode, _ := t.getConnection(addr)
	if peerNode != nil {
		return peerNode, nil
	}
	conn, err := net.Dial(t.ListenAddr.Network(), addr)
	fmt.Println(addr, err)
	peerNode = t.addConnection(conn)

	t.write(peerNode, metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection with %s", addr)
	}

	if t.ShouldClientHandleConn {
		go t.handleConnection(conn)
	}

	return peerNode, nil
}

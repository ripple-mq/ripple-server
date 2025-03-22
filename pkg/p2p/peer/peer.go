package peer

import "net"

type Peer interface {
	GetAddress() net.Addr
	GetConnection() net.Conn
	Close() error
}

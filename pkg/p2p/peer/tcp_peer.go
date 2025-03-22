package peer

import "net"

type TCPPeer struct {
	Conn net.Conn
}

func (p *TCPPeer) GetAddress() net.Addr {
	return p.Conn.RemoteAddr()
}

func (p *TCPPeer) GetConnection() net.Conn {
	return p.Conn
}

func (p *TCPPeer) Close() error {
	return p.Conn.Close()
}

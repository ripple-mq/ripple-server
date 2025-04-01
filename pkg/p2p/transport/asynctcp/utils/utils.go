package utils

import (
	"fmt"
	"net"
	"syscall"
)

func SockaddrToString(sa syscall.Sockaddr) string {
	switch addr := sa.(type) {
	case *syscall.SockaddrInet4:
		ip := net.IP(addr.Addr[:])
		return fmt.Sprintf("%s:%d", ip.String(), addr.Port)
	case *syscall.SockaddrInet6:
		ip := net.IP(addr.Addr[:])
		return fmt.Sprintf("[%s]:%d", ip.String(), addr.Port)
	case *syscall.SockaddrUnix:
		return addr.Name
	default:
		return "Unknown sockaddr type"
	}
}

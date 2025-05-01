package server

import (
	"net"

	"github.com/charmbracelet/log"
	pb "github.com/ripple-mq/ripple-server/server/exposed/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedBootstrapServerServer
}

// BootstrapServer represents a server for bootstrapping connections.
// It holds the address, listener, and gRPC server instances.
type BootstrapServer struct {
	Addr     net.Addr
	listener *net.Listener
	server   *grpc.Server
}

// NewBootstrapServer creates and returns a new BootstrapServer instance.
func NewBootstrapServer(addr string) (*BootstrapServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	pb.RegisterBootstrapServerServer(s, Server{})
	reflection.Register(s)
	return &BootstrapServer{
		Addr:     listener.Addr(),
		listener: &listener,
		server:   s,
	}, nil
}

// Listen starts the bootstrap server and begins serving metadata requests.
func (t *BootstrapServer) Listen() error {
	go func() {
		log.Infof("started bootstrap server metadata service, listening on port: %s", t.Addr)
		if err := t.server.Serve(*t.listener); err != nil {
			log.Fatalf("failed to listen on port %s", t.Addr)
		}
	}()
	return nil
}

// Stop stops gRPC server
func (t *BootstrapServer) Stop() {
	t.server.Stop()
}

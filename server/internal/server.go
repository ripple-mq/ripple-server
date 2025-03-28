package server

import (
	"net"

	"github.com/charmbracelet/log"
	pb "github.com/ripple-mq/ripple-server/server/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedInternalServiceServer
}

// InternalServer represents a server for bootstrapping connections.
// It holds the address, listener, and gRPC server instances.
type InternalServer struct {
	Addr     net.Addr
	listener *net.Listener
	server   *grpc.Server
}

// NewInternalServer initializes and returns a new internal server that listens on the specified address.
// It creates a TCP listener, sets up a gRPC server, and registers the internal service.
func NewInternalServer(addr string) (*InternalServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	pb.RegisterInternalServiceServer(s, Server{})
	reflection.Register(s)
	return &InternalServer{
		Addr:     listener.Addr(),
		listener: &listener,
		server:   s,
	}, nil
}

// Listen starts the internal server to begin listening for incoming gRPC requests.
// It runs the server in a separate goroutine and logs any errors during the server operation.
func (t *InternalServer) Listen() error {
	go func() {
		log.Infof("started bootstrap server metadata service, listening on port: %s", t.Addr)
		if err := t.server.Serve(*t.listener); err != nil {
			log.Fatalf("failed to listen on port %s", t.Addr)
		}
	}()
	return nil
}

// Stop stops the internal server gracefully, terminating all ongoing connections
func (t *InternalServer) Stop() {
	t.server.Stop()
}

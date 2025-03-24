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

type InternalServer struct {
	Addr     net.Addr
	listener *net.Listener
	server   *grpc.Server
}

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

func (t *InternalServer) Listen() error {
	go func() {
		log.Infof("started bootstrap server metadata service, listening on port: %s", t.Addr)
		if err := t.server.Serve(*t.listener); err != nil {
			log.Fatalf("failed to listen on port %s", t.Addr)
		}
	}()
	return nil
}

func (t *InternalServer) Stop() {
	t.server.Stop()
}

package server

import (
	"log"

	broker "github.com/ripple-mq/ripple-server/server/exposed"
	bootsrrap "github.com/ripple-mq/ripple-server/server/internal"
)

type Server struct {
	brokerServer    *broker.BootstrapServer
	bootstrapServer *bootsrrap.InternalServer
}

func NewServer(baddr string, eaddr string) *Server {
	brokerServer, _ := broker.NewBootstrapServer(eaddr)
	bootstrapServer, _ := bootsrrap.NewInternalServer(baddr)

	return &Server{brokerServer, bootstrapServer}
}

func (t *Server) Listen() {
	err := t.bootstrapServer.Listen()
	if err != nil {
		log.Fatalf("failed to spin up bostrap server: %v", err)
	}
	err = t.brokerServer.Listen()
	if err != nil {
		log.Fatalf("failed to spin up broker server: %v", err)
	}
}

func (t *Server) Stop() {
	t.bootstrapServer.Stop()
	t.brokerServer.Stop()
}

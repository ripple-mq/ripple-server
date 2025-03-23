package server

import (
	"log"

	bootsrrap "github.com/ripple-mq/ripple-server/server/bootstrap"
	broker "github.com/ripple-mq/ripple-server/server/exposed"
)

type Server struct {
	brokerServer    *broker.BrokerServer
	bootstrapServer *bootsrrap.BootstrapServer
}

func NewServer(baddr string, eaddr string) *Server {
	brokerServer, _ := broker.NewBrokerServer(eaddr)
	bootstrapServer, _ := bootsrrap.NewBoostrapServer(baddr)

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

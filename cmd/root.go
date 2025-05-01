package cmd

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	"github.com/ripple-mq/ripple-server/pkg/utils/env"
	"github.com/ripple-mq/ripple-server/pkg/utils/pen"
	"github.com/ripple-mq/ripple-server/server"
)

const (
	minPort = 1024
	maxPort = 49150
)

func RandLocalAddr() string {
	randomNumber := rand.Intn(maxPort-minPort) + minPort
	return fmt.Sprintf(":%d", randomNumber)
}

func printBanner() {
	fmt.Print(`
	██████╗ ██╗██████╗ ██████╗ ██╗     ███████╗
	██╔══██╗██║██╔══██╗██╔══██╗██║     ██╔════╝
	██████╔╝██║██████╔╝██████╔╝██║     █████╗  
	██╔══██╗██║██╔═══╝ ██╔═══╝ ██║     ██╔══╝  
	██║  ██║██║██║     ██║     ███████╗███████╗
	╚═╝  ╚═╝╚═╝╚═╝     ╚═╝     ╚══════╝╚══════╝
	`)
	fmt.Println()
}

func Execute() {
	pen.SpinWheel("Starting server... ", "Finished 🎉")

	cfg := config.Conf
	l := pen.Loader("Starting ripple server... ")
	go func() {
		addr := fmt.Sprintf("%s:%d", env.Get("ASYNC_TCP_IPv4", "127.0.0.1"), 6060)
		log.Info("Profiling started")
		log.Info(http.ListenAndServe(addr, nil))
	}()
	log.Info(cfg)

	internal := fmt.Sprintf("%s:%s", env.Get("ASYNC_TCP_IPv4", "127.0.0.1"), config.Conf.Server.Internal_grpc_addr)
	bootstrap := fmt.Sprintf("%s:%s", env.Get("ASYNC_TCP_IPv4", "127.0.0.1"), config.Conf.Server.Exposed_grpc_addr)
	s := server.NewServer(internal, bootstrap)
	s.Listen()
	time.Sleep(2 * time.Second)
	printBanner()

	pen.Complete(l, "ripple server initiated successfully 🎉")
	select {}
}

package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
	"github.com/ripple-mq/ripple-server/pkg/utils/env"
	"github.com/ripple-mq/ripple-server/pkg/utils/pen"
	"github.com/ripple-mq/ripple-server/server"
)

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

	go func() {
		addr := fmt.Sprintf("%s:%d", env.Get("ASYNC_TCP_IPv4", "127.0.0.1"), 6060)
		log.Info("Profiling started")
		http.ListenAndServe(addr, nil)
	}()

	internal := fmt.Sprintf("%s:%s", env.Get("ASYNC_TCP_IPv4", "127.0.0.1"), config.Conf.Server.Internal_grpc_addr)
	bootstrap := fmt.Sprintf("%s:%s", env.Get("ASYNC_TCP_IPv4", "127.0.0.1"), config.Conf.Server.Exposed_grpc_addr)
	s := server.NewServer(internal, bootstrap)
	s.Listen()
	printBanner()

	pen.SpinWheel("Starting ripple server... ", "ripple server initiated successfully 🎉")
	select {}
}

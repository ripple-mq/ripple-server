package main

import (
	"github.com/ripple-mq/ripple-server/cmd"
	"github.com/ripple-mq/ripple-server/pkg/utils/pen"
)

func main() {
	pen.InitLog()
	cmd.Execute()
}

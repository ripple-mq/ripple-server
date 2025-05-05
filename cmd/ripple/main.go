package main

import (
	"github.com/ripple-mq/ripple-server/pkg/utils/pen"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ripple",
		Short: "Ripple CLI",
	}

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the ripple server",
		Run: func(cmd *cobra.Command, args []string) {
			pen.InitLog()
			cmd.Execute()
		},
	}

	rootCmd.AddCommand(startCmd)
	rootCmd.Execute()
}

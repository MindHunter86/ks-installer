package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var (
	cmdServer = &cobra.Command{
		Use: "server [command]",
		Short: "Command group for server controll",
		Args: cobra.MinimumNArgs(1),
	}
	cmdServerServe = &cobra.Command{
		Use: "serve",
		Short: "Start server serving",
		Run: func(cmd *cobra.Command, args []string) {
			log.Print("Hey, i'm log")
		},
	}
)

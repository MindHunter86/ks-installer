package cmd

import "github.com/spf13/cobra"

var (
	cmdServer = &cobra.Command{
		Use: "server [command]",
		Short: "Command group for server controll",
		Args: cobra.MinimumNArgs(1)
	}
	cmdServerServe = &cobra.Command{}
)

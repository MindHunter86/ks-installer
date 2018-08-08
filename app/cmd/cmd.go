package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
		"os"
	"log"
)

var (
	globViper *viper.Viper

	flConfig string
	flRestore bool

	KsInstaller = &cobra.Command{
		Use: "",
		Short: "",
		Long: "",

		PersistentPreRun: func(cmd *cobra.Command, args []string) {

			viper.SetConfigName("ks-installer")

			viper.SetConfigType("yaml")

			viper.AddConfigPath("/etc/ks-installer")
			viper.AddConfigPath("/etc/sysconfig/ks-installer")
			viper.AddConfigPath("$HOME/.ks-installer")
			viper.AddConfigPath("./extras")

			if e := viper.ReadInConfig(); e != nil {
				log.Printf("could not parse the configuration file; error: %e", e)
				os.Exit(1)
			}

			globViper = viper.GetViper()
		},

		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
)

func init() {

	cmdServerServe.Flags().StringVarP(&flConfig, "config", "c","", "specify config file path")
	cmdServerServe.Flags().BoolVar(&flRestore, "restore-mode", false, "start data restore mode")

	viper.BindPFlag("base.raft.restore_mode", cmdServerServe.Flags().Lookup("restore-mode"))

	cmdServer.AddCommand(cmdServerServe)
	KsInstaller.AddCommand(cmdServer)
}

func GetViper() *viper.Viper {
	return globViper
}

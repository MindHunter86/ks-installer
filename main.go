package main

import "os"
import "flag"
// import "bitbucket.org/mh00net/ks-installer/client"
// import "bitbucket.org/mh00net/ks-installer/installer"
import "bitbucket.org/mh00net/ks-installer/core"
import "bitbucket.org/mh00net/ks-installer/core/config"
import "github.com/rs/zerolog"


var log zerolog.Logger
const argHelp = `
usage: ks-installer <command> [<args>]
command list:
	* master - command group for master instance
		* serve - starting master serve

	* server - command group for server management
		* add - add server for future reinstallation
		* install - command for gathering Ethernet information and starting client event loop. Used by anaconda in %pre scriptlet
		* setup - starting base wrapper for puppet agent. Used by clean OS for first puppet runs
`

func main() {

	var err error

	fgMasterServeSet := flag.NewFlagSet("serve", flag.ExitOnError)
	masterServeConfig := fgMasterServeSet.String("config", "./config.yml", "filepath to the configuration file")

	fgServerAddSet := flag.NewFlagSet("add", flag.ExitOnError)
	// serverAddHostname := fgServerAddSet.String("hostname", "", "server hostname")
	// serverAddMAC := fgServerAddSet.String("mac", "ff:ff:ff:ff:ff:ff", "server MAC addr of the one of links")

	fgServerInstall := flag.NewFlagSet("install", flag.ExitOnError)
	// serverInstallMaster := fgServerInstall.String("master", "", "Master's IPv4 address")

	fgServerSetup := flag.NewFlagSet("setup", flag.ExitOnError)
	// serverSetupTest := fgServerSetup.String("test", "bar", "test foo=bar")

	// log initialization:
	zerolog.ErrorFieldName = "ERROR"
	log = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr }).With().Timestamp().Logger()

	// parse all given arguments:
	switch {
		case len(os.Args) <= 2: log.Print(argHelp); os.Exit(2)
		case os.Args[1] == "master" && os.Args[2] == "serve": fgMasterServeSet.Parse(os.Args[3:])
		case os.Args[1] == "server" && os.Args[2] == "add": fgServerAddSet.Parse(os.Args[3:])
		case os.Args[1] == "server" && os.Args[2] == "add": fgServerInstall.Parse(os.Args[3:])
		case os.Args[1] == "server" && os.Args[2] == "add": fgServerSetup.Parse(os.Args[3:])
		default: log.Print(argHelp); os.Exit(2)
	}

	switch {
		case fgMasterServeSet.Parsed():
			if len(*masterServeConfig) == 0 {
				log.Fatal().Msg("The Config argument cannot be empty!"); os.Exit(2) }

			// read and parse config:
			if _,e := os.Stat(*masterServeConfig); e != nil {
				log.Fatal().Err(e).Msg("Could not stat configuration file!")
				os.Exit(1) }
			cfg,e := new(config.CoreConfig).Parse(*masterServeConfig); if e != nil {
				log.Fatal().Err(e).Msg("Could not successfully complete configuration file parsing!")
				os.Exit(1) }

			// log configuration:
			switch cfg.Base.Log_Level {
				case "off": zerolog.SetGlobalLevel(zerolog.NoLevel)
				case "debug": zerolog.SetGlobalLevel(zerolog.DebugLevel)
				case "info": zerolog.SetGlobalLevel(zerolog.InfoLevel)
				case "warn": zerolog.SetGlobalLevel(zerolog.WarnLevel)
				case "error": zerolog.SetGlobalLevel(zerolog.ErrorLevel)
				case "fatal": zerolog.SetGlobalLevel(zerolog.FatalLevel)
				case "panic": zerolog.SetGlobalLevel(zerolog.PanicLevel) }

			// core initialization:
			core,e := new(core.Core).SetLogger(&log).SetConfig(cfg).Construct(); if e != nil {
				log.Fatal().Err(e).Msg("Could not successfully complete the Costruct method!")
				os.Exit(1) }

			err = core.Bootstrap()

		case fgServerAddSet.Parsed():
		case fgServerInstall.Parsed():
		case fgServerSetup.Parsed():
		default: log.Print("fuckit")
	}


	// check subprogram errors and exit:
	if err != nil {
		log.Fatal().Err(err).Msg("Abnormal program result!"); os.Exit(1)
	} else {
		log.Debug().Msg("Program completed successfully!"); os.Exit(0) }
}

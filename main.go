package main

import "os"
import "time"

// import "bitbucket.org/mh00net/ks-installer/client"
// import "bitbucket.org/mh00net/ks-installer/installer"
import "bitbucket.org/mh00net/ks-installer/core"
import "bitbucket.org/mh00net/ks-installer/core/config"
import "github.com/rs/zerolog"
import "gopkg.in/urfave/cli.v1"

var log zerolog.Logger

func main() {

	// log initialization:
	zerolog.ErrorFieldName = "ERROR"
	log = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr}).With().Timestamp().Logger()

	// define all commands && flags:
	app := cli.NewApp()
	app.Name = "ks-installer"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Vadimka Komissarov",
			Email: "v.komissarov@corp.mail.ru, vadimka_kom@mail.ru"}}
	app.Copyright = "(c) 2018 Mindhunter and CO"
	app.Usage = "Kickstart install manager for M***Ru PortalAdminz"

	app.Commands = []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "command for server management",
			Subcommands: []cli.Command{
				{
					Name:    "serve",
					Aliases: []string{"s"},
					Usage:   "start serving",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "config, c",
							Usage:  "Load configuration file for server from `FILE`",
							Value:  "./extras/config.yml",
							EnvVar: "SERVER_CONFIG",
						},
					},
					Action: func(c *cli.Context) error {

						// stat() and parse configuration file:
						if _, e := os.Stat(c.String("config")); e != nil {
							return e
						}
						cfg, e := new(config.CoreConfig).Parse(c.String("config"))
						if e != nil {
							return e
						}

						// global logger configuration:
						switch cfg.Base.Log_Level {
						case "off":
							zerolog.SetGlobalLevel(zerolog.NoLevel)
						case "debug":
							zerolog.SetGlobalLevel(zerolog.DebugLevel)
						case "info":
							zerolog.SetGlobalLevel(zerolog.InfoLevel)
						case "warn":
							zerolog.SetGlobalLevel(zerolog.WarnLevel)
						case "error":
							zerolog.SetGlobalLevel(zerolog.ErrorLevel)
						case "fatal":
							zerolog.SetGlobalLevel(zerolog.FatalLevel)
						case "panic":
							zerolog.SetGlobalLevel(zerolog.PanicLevel)
						}

						// core initialization:
						appCore, e := new(core.Core).SetLogger(&log).SetConfig(cfg).Construct()
						if e != nil {
							return e
						}

						// core bootstrap:
						return appCore.Bootstrap()
					},
				},
			},
		},
		{
			Name:    "host",
			Aliases: []string{"ho"},
			Usage:   "command for host management",
			Subcommands: []cli.Command{
				{
					Name:     "add",
					Aliases:  []string{"a"},
					Usage:    "add host for future reinstallation",
					Category: "host",
					Action: func(c *cli.Context) error {
						return nil
					},
				},
				{
					Name:     "install",
					Aliases:  []string{"i"},
					Usage:    "command for gathering Ethernet information and starting client event loop. Used by anaconda in %pre scriptlet",
					Category: "host",
					Action: func(c *cli.Context) error {
						return nil
					},
				},
				{
					Name:     "setup",
					Aliases:  []string{"s"},
					Usage:    "starting base wrapper for puppet agent. Used by clean OS for first puppet runs",
					Category: "host",
					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
		},
	}

	// parse all given arguments:
	if e := app.Run(os.Args); e != nil {
		log.Fatal().Err(e).Msg("Could not run the App!")
	}
}

func old_main() {

	// fgMasterServeSet := flag.NewFlagSet("serve", flag.ExitOnError)
	// masterServeConfig := fgMasterServeSet.String("config", "./config.yml", "filepath to the configuration file")

	// fgServerAddSet := flag.NewFlagSet("add", flag.ExitOnError)
	// serverAddHostname := fgServerAddSet.String("hostname", "", "server hostname")
	// serverAddMAC := fgServerAddSet.String("mac", "ff:ff:ff:ff:ff:ff", "server MAC addr of the one of links")

	// fgServerInstall := flag.NewFlagSet("install", flag.ExitOnError)
	// serverInstallMaster := fgServerInstall.String("master", "", "Master's IPv4 address")

	// fgServerSetup := flag.NewFlagSet("setup", flag.ExitOnError)
	// serverSetupTest := fgServerSetup.String("test", "bar", "test foo=bar")
}

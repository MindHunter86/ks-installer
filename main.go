package main

import "os"

import "github.com/rs/zerolog"
import (
		"github.com/MindHunter86/ks-installer/app/cmd"
)
// import "gopkg.in/urfave/cli.v1/altsrc"

var log zerolog.Logger

func main() {

	// log initialization:
	zerolog.ErrorFieldName = "ERROR"
	log = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr}).With().Timestamp().Logger()

	// define all commands && flags (new plugin):
	if e := cmd.KsInstaller.Execute(); e != nil {
		log.Error().Err(e).Msg("")
		os.Exit(1)
	}

//	// define all commands && flags:
//	app := cli.NewApp()
//	app.Name = "ks-installer"
//	app.Version = "0.0.1"
//	app.Compiled = time.Now()
//	app.Authors = []cli.Author{
//		{
//			Name:  "Vadimka Komissarov",
//			Email: "v.komissarov@corp.mail.ru, vadimka_kom@mail.ru"}}
//	app.Copyright = "(c) 2018 Mindhunter and CO"
//	app.Usage = "Kickstart install manager for M***Ru PortalAdminz"
//
//	app.Commands = []cli.Command{
//		{
//			Name:    "server",
//			Aliases: []string{"s"},
//			Usage:   "command for server management",
//			Subcommands: []cli.Command{
//				{
//					Name:    "serve",
//					Aliases: []string{"s"},
//					Usage:   "start serving",
//					Flags: []cli.Flag{
//						cli.StringFlag{
//							Name:   "config, c",
//							Usage:  "Load configuration file for server from `FILE`",
//							Value:  "./extras/config.yml",
//							EnvVar: "SERVER_CONFIG",
//						},
//						cli.BoolFlag{
//							Name:  "master, m",
//							Usage: "Force RAFT cluster bootstrap. Use it carefully!",
//						},
//					},
//					Action: func(c *cli.Context) error {
//
//						// stat() and parse configuration file:
//						if _, e := os.Stat(c.String("config")); e != nil {
//							return e
//						}
//						cfg, e := new(config.CoreConfig).Parse(c.String("config"))
//						if e != nil {
//							return e
//						}
//
//						// global logger configuration:
//						switch cfg.Base.Log_Level {
//						case "off":
//							zerolog.SetGlobalLevel(zerolog.NoLevel)
//						case "debug":
//							zerolog.SetGlobalLevel(zerolog.DebugLevel)
//						case "info":
//							zerolog.SetGlobalLevel(zerolog.InfoLevel)
//						case "warn":
//							zerolog.SetGlobalLevel(zerolog.WarnLevel)
//						case "error":
//							zerolog.SetGlobalLevel(zerolog.ErrorLevel)
//						case "fatal":
//							zerolog.SetGlobalLevel(zerolog.FatalLevel)
//						case "panic":
//							zerolog.SetGlobalLevel(zerolog.PanicLevel)
//						}
//
//						// core initialization:
//						appCore, e := new(core.Core).SetLogger(&log).SetConfig(cfg).Construct()
//						if e != nil {
//							return e
//						}
//
//						// core bootstrap:
//						return appCore.Bootstrap(c.Bool("master"))
//					},
//				},
//			},
//		},
//		{
//			Name:    "host",
//			Aliases: []string{"ho"},
//			Usage:   "command for host management",
//			Subcommands: []cli.Command{
//				{
//					Name:     "add",
//					Aliases:  []string{"a"},
//					Usage:    "add host for future reinstallation",
//					Category: "host",
//					Action: func(c *cli.Context) error {
//						return nil
//					},
//				},
//				{
//					Name:     "install",
//					Aliases:  []string{"i"},
//					Usage:    "command for gathering Ethernet information and starting client event loop. Used by anaconda in %pre scriptlet",
//					Category: "host",
//					Action: func(c *cli.Context) error {
//						return nil
//					},
//				},
//				{
//					Name:     "setup",
//					Aliases:  []string{"s"},
//					Usage:    "starting base wrapper for puppet agent. Used by clean OS for first puppet runs",
//					Category: "host",
//					Action: func(c *cli.Context) error {
//						return nil
//					},
//				},
//			},
//		},
//	}
//
//	// parse all given arguments:
//	if e := app.Run(os.Args); e != nil {
//		log.Fatal().Err(e).Msg("Could not run the App!")
//	}
}

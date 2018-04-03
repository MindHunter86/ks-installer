package main

import "os"
import "flag"
import "mailru/tgrmalerter/core"
import "mailru/tgrmalerter/core/config"
import "github.com/rs/zerolog"

var parsedConfigFile string
const defaultConfigFile string = "./config.yml"

func init() {
	flag.StringVar(&parsedConfigFile, "c", defaultConfigFile, "path to configuration file (default: ./config.yml)")
}

func main() {
	// parse all mservice given arguments:
	flag.Parse()

	// log initialization:
	zerolog.ErrorFieldName = "ERROR"
	log := zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr }).With().Timestamp().Logger()
	log.Debug().Msg("Logger has been successfully initialized!")

	// parse config:
	if _,e := os.Stat(parsedConfigFile); e != nil {
		log.Fatal().Err(e).Msg("Could not stat configuration file!")
		os.Exit(1) }
	cfg,e := new(config.CoreConfig).Parse(parsedConfigFile); if e != nil {
		log.Fatal().Err(e).Msg("Could not successfully complete ( /a/the) configuration file parsing!")
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

	// core bootstrap:
	if e = core.Bootstrap(); e != nil {
		log.Error().Err(e).Msg("Runtime error! Bootstrap or Destruct methods have been failed!")
		os.Exit(1) }

	// main() footer:
	log.Info().Msg("Core has been successfully stopped and destroyed!")
	os.Exit(0)
}

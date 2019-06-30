package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig"
)

type command string
type commandFunc func() error

const (
	cmdHelp        command = "help"
	cmdShare       command = "share"
	cmdSync        command = "sync"
	cmdWriteConfig command = "write-config"
)

var cmdFuncs = map[command]commandFunc{
	cmdShare:       execShare,
	cmdSync:        execSync,
	cmdWriteConfig: execWriteSampleConfig,
}

var (
	cfg = struct {
		Config         string `flag:"config,c" default:"config.yaml" description:"Configuration file location"`
		Force          bool   `flag:"force,f" default:"false" description:"Force operation"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("cloudbox %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}

	if dir, err := homedir.Expand(cfg.Config); err != nil {
		log.WithError(err).Fatal("Unable to expand config path")
	} else {
		cfg.Config = dir
	}
}

func main() {
	cmd := cmdHelp
	if len(rconfig.Args()) > 1 {
		cmd = command(rconfig.Args()[1])
	}

	var cmdFunc commandFunc = execHelp
	if f, ok := cmdFuncs[cmd]; ok {
		cmdFunc = f
	}

	log.WithField("version", version).Info("cloudbox started")

	if err := cmdFunc(); err != nil {
		log.WithError(err).Fatal("Command execution failed")
	}
}

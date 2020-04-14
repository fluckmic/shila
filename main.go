package main

import (
	"fmt"
	"os"
	"shila/config"
	"shila/logging"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	var cfg config.Config
	var log *logging.Logger
	var err error

	// Load the initialization
	// TODO: Load config from a file as an alternative.
	if err = cfg.InitDefault(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	// Set up the logging
	if log, err = logging.New(cfg.Logging); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	log.Infoln("Shila initialization successful.")
	log.Debugf("%T : %+v\n", cfg.Logging, cfg.Logging)
	log.Debugf("%T : %+v\n", cfg.KernelEndpoint, cfg.KernelEndpoint)

	return 0
}

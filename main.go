package main

import (
	"fmt"
	"os"
	"shila/config"
	"shila/logging"
)

var (
	cfg config.Config
	log *logging.Logger
	err error
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	// Load the initialization
	// TODO: Load config from a file as an alternative.
	if err = cfg.InitDefault(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	// Set up the logging
	if log, err = logging.New(cfg.Logging); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	log.Infoln("Shila initialization successful.")
	log.Debugf("%T : %+v\n", cfg.Logging, cfg.Logging)

	return 0
}

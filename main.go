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
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	// Load the initialization
	// TODO: Load config from a file as an alternative.
	if err := cfg.InitDefault(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	// Set up the logging
	// TODO: Error handling.
	log = logging.New(cfg.Logging)

	return 0
}

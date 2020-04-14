package main

import (
	"fmt"
	"os"
	"os/signal"
	"shila/config"
	"shila/fatal"
	"shila/kersi"
	"shila/logging"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	var cfg config.Config
	var log *logging.Logger
	var err error

	// Initialize termination functionality
	fatal.Init()

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
	// TODO: defer log.TearDown: properly close logging afterwards.

	log.Infoln("Shila initialization successful.")
	log.Debugf("%T : %+v\n", cfg.Logging, cfg.Logging)
	log.Debugf("%T : %+v\n", cfg.KernelSide, cfg.KernelSide)
	log.Debugf("%T : %+v\n", cfg.KernelEndpoint, cfg.KernelEndpoint)

	// Create and setup the kernel side
	var kernelSide = kersi.New(cfg)
	if err = kernelSide.Setup(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	defer kernelSide.CleanUp()

	returnCode := waitForTeardown()

	// Clean up which cannot be done with defer.

	return returnCode
}

func waitForTeardown() int {

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, os.Kill)

	select {
	case <-c:
		return 0
	case <-fatal.ShutdownChan():
		return 0
	case <-fatal.FatalChan():
		return 1
	}
}

package main

import (
	"os"
	"shila/config"
	"shila/kersi"
	"shila/log"
	"shila/shutdown"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	defer log.Info.Println("Shutdown complete.")

	var cfg config.Config
	var err error

	// Initialize logging functionality
	log.Init()

	// Initialize termination functionality
	shutdown.Init()

	log.Info.Println("Setup started...")

	// Load the initialization
	// TODO: Load config from a file as an alternative.
	if err = cfg.InitDefault(); err != nil {
		log.Error.Fatalln("Fatal error -", err.Error())
		return 1
	}
	log.Verbose.Println("Configuration loaded.")

	// Create and setup the kernel side
	var kernelSide = kersi.New(cfg)
	if err = kernelSide.Setup(); err != nil {
		log.Error.Fatalln("Fatal error -", err.Error())
	}
	defer kernelSide.CleanUp()
	log.Verbose.Println("Setup kernel side done.")

	log.Info.Println("Setup done, starting machinery..")

	log.Info.Println("Machinery up and running.")
	returnCode := waitForTeardown()

	// Clean up which cannot be done with defer.

	return returnCode
}

func waitForTeardown() int {
	select {
	case <-shutdown.OrderlyChan():
		return 0
	case <-shutdown.FatalChan():
		return 1
	}
}

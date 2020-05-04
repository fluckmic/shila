package main

import (
	"os"
	"shila/config"
	"shila/kernelSide"
	"shila/log"
	"shila/shutdown"
	"shila/stumps"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	defer log.Info.Println("Shutdown complete.")

	// TODO: Return if not run as root.

	// TODO: Encourage user to run shila in separate namespace for ingress and egress

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
	}

	// Create and setup the kernel side
	var kernelSide = kernelSide.New(cfg)
	if err = kernelSide.Setup(); err != nil {
		log.Error.Fatalln("Fatal error -", err.Error())
	}
	defer kernelSide.CleanUp()

	// TODO: Create and setup the network side

	// TODO: Create and setup the core
	stumps.Setup(kernelSide)

	log.Info.Println("Setup done, starting machinery..")

	// TODO: Run the machinery.
	_ = stumps.Start()

	if err = kernelSide.Start(); err != nil {
		log.Error.Fatalln("Fatal error -", err.Error())
	}

	log.Info.Println("Machinery up and running.")

	returnCode := waitForTeardown()

	// TODO: Clean everything up

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

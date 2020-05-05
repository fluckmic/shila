package main

import (
	"os"
	"shila/config"
	"shila/kernelSide"
	"shila/log"
	"shila/networkSide"
	"shila/shila/connection"
	"shila/shutdown"
	"shila/workingSide"
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
	kernelSide := kernelSide.New(cfg)
	if err = kernelSide.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the kernel side - ", err.Error())
	}
	log.Info.Println("Kernel side setup successfully.")
	defer kernelSide.CleanUp()

	// Create and setup the network side
	networkSide := networkSide.New(cfg)
	if err = networkSide.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the network side - ", err.Error())
	}
	log.Info.Println("Network side setup successfully.")
	defer networkSide.CleanUp()

	// Create the mapping holding the connections
	connections := connection.NewMapping(kernelSide, networkSide)

	// Create and setup the working side
	workingSide := workingSide.New(cfg, kernelSide, networkSide, connections)
	if err := workingSide.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the working side - ", err.Error())
	}
	log.Info.Println("Working side setup successfully.")
	defer workingSide.CleanUp()

	log.Info.Println("Setup done, starting machinery..")

	if err = workingSide.Start(); err != nil {
		log.Error.Fatalln("Unable to start the working side - ", err.Error())
	}

	if err = networkSide.Start(); err != nil {
		log.Error.Fatalln("Unable to start the network side - ", err.Error())
	}

	if err = kernelSide.Start(); err != nil {
		log.Error.Fatalln("Unable to start the kernel side - ", err.Error())
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

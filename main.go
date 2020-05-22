package main

import (
	"os"
	"shila/config"
	"shila/core/connection"
	"shila/core/netflow"
	"shila/core/shila"
	"shila/kernelSide"
	"shila/log"
	"shila/networkSide"
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

	// Create the channel used to announce new traffic channels
	trafficChannelAnnouncements := make(chan shila.PacketChannelAnnouncement)

	// Create and setup the kernel side
	kernel := kernelSide.New(cfg, trafficChannelAnnouncements)
	if err = kernel.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the kernel side - ", err.Error())
	}
	log.Info.Println("Kernel side setup successfully.")
	defer kernel.CleanUp()

	// Create and setup the network side
	network := networkSide.New(cfg, trafficChannelAnnouncements)
	if err = network.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the network side - ", err.Error())
	}
	log.Info.Println("Network side setup successfully.")
	defer network.CleanUp()

	// Create the mapping holding the network addresses
	routing := netflow.NewRouter()

	// TODO. ############## Testing ##############
	key := "(10.7.0.9:2727)"
	path := networkSide.Generator{}.NewPath("")
	dstAddr := networkSide.Generator{}.NewAddress("192.168.34.189:2727")
	srcAddr := networkSide.Generator{}.NewEmptyAddress()
	// TODO. ############## Testing ##############

	routing.InsertFromIPAddressPortKey(shila.IPAddressPortKey(key), shila.NetFlow{srcAddr, path, dstAddr})

	// Create the mapping holding the connections
	connections := connection.NewMapping(kernel, network, routing)

	// Create and setup the working side
	workingSide := workingSide.New(cfg, connections, trafficChannelAnnouncements)
	if err := workingSide.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the working side - ", err.Error())
	}
	log.Info.Println("Working side setup successfully.")
	defer workingSide.CleanUp()

	log.Info.Println("Setup done, starting machinery..")

	if err = workingSide.Start(); err != nil {
		log.Error.Fatalln("Unable to start the working side - ", err.Error())
	}

	if err = network.Start(); err != nil {
		log.Error.Fatalln("Unable to start the network side - ", err.Error())
	}

	if err = kernel.Start(); err != nil {
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

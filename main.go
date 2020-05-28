package main

import (
	"os"
	"shila/core/connection"
	"shila/core/netflow"
	"shila/core/shila"
	"shila/kernelSide"
	"shila/log"
	"shila/networkSide"
	"shila/networkSide/network"
	"shila/shutdown"
	"shila/workingSide"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	defer log.Info.Println("Shutdown complete.")

	// TODO: Return if not run as root. (https://github.com/fluckmic/shila/issues/4)
	// TODO: Encourage user to run separate namespace for ingress and egress (https://github.com/fluckmic/shila/issues/5)

	var err error

	// Initialize logging functionality
	log.Init()

	// Initialize termination functionality
	shutdown.Init()

	log.Info.Println("Setup started...")

	// Create the channel used to announce new traffic channels
	trafficChannelAnnouncements := make(chan shila.PacketChannelAnnouncement)

	// Create and setup the kernelSide side
	kernelSide := kernelSide.New(trafficChannelAnnouncements)
	if err = kernelSide.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the kernel side - ", err.Error())
	}
	log.Info.Println("Kernel side setup successfully.")
	defer kernelSide.CleanUp()

	// Create and setup the network side
	networkSide := networkSide.New(trafficChannelAnnouncements)
	if err = networkSide.Setup(); err != nil {
		log.Error.Fatalln("Unable to setup the network side - ", err.Error())
	}
	log.Info.Println("Network side setup successfully.")
	defer networkSide.CleanUp()

	// Create the mapping holding the network addresses
	routing := netflow.NewRouter()

	// TODO. ############## Testing ##############
	key := "(10.7.0.9:2727)"
	path := network.PathGenerator{}.NewEmpty()
	dstAddr := network.AddressGenerator{}.New("192.168.22.131:2727")
	srcAddr := network.AddressGenerator{}.NewEmpty()
	// TODO. ############## Testing ##############

	routing.InsertFromIPAddressPortKey(shila.IPAddressPortKey(key), shila.NetFlow{srcAddr, path, dstAddr})

	// Create the mapping holding the connections
	connections := connection.NewMapping(kernelSide, networkSide, routing)

	// Create and setup the working side
	workingSide := workingSide.New(connections, trafficChannelAnnouncements)
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
		log.Error.Fatalln("Unable to start the kernelSide side - ", err.Error())
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

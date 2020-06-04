package main

import (
	"os"
	"shila/core/connection"
	"shila/core/netflow"
	"shila/core/shila"
	"shila/kernelSide"
	"shila/log"
	"shila/networkSide"
	"shila/shutdown"
	"shila/workingSide"
)

const (
	SuccessCode = 0
	ErrorCode   = 1
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

	// Create the channel used to announce new traffic channels and possible issues within endpoints.
	trafficChannelPubs := make(shila.PacketChannelPubChannel)
	endpointIssues := make(shila.EndpointIssuePubChannel)

	// Create and setup the kernelSide side
	kernelSide := kernelSide.New(trafficChannelPubs, endpointIssues)
	if err = kernelSide.Setup(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to setup kernel side.").Error())
		return ErrorCode
	}
	log.Info.Println("Kernel side setup successfully.")
	defer kernelSide.CleanUp()

	// Create and setup the network side
	networkSide := networkSide.New(trafficChannelPubs, endpointIssues)
	if err = networkSide.Setup(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to setup network side.").Error())
		return ErrorCode
	}
	log.Info.Println("Network side setup successfully.")
	defer networkSide.CleanUp()

	// Create the mapping holding the network addresses
	routing := netflow.NewRouter()

	// Create the mapping holding the connections
	connections := connection.NewMapping(kernelSide, networkSide, routing)

	// Create and setup the working side
	workingSide := workingSide.New(connections, trafficChannelPubs, endpointIssues)
	if err := workingSide.Setup(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to setup working side.").Error())
		return ErrorCode
	}
	log.Info.Println("Working side setup successfully.")
	defer workingSide.CleanUp()

	log.Info.Println("Setup done, starting machinery..")

	if err = workingSide.Start(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to start working side.").Error())
		return ErrorCode
	}
	if err = networkSide.Start(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to start network side.").Error())
		return ErrorCode
	}
	if err = kernelSide.Start(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to start the kernel side.").Error())
		return ErrorCode
	}

	log.Info.Println("Machinery up and running.")

	returnCode := waitForTeardown()
	return returnCode
}

func waitForTeardown() int {
	select {
	case <-shutdown.OrderlyChan():
		return SuccessCode
	case <-shutdown.FatalChan():
		return ErrorCode
	}
}

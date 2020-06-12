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

	log.Verbose.Println("Setup started...")

	// Create the channel used to announce new traffic channels and possible issues within endpoints.
	trafficChannelPubs := shila.PacketChannelPubChannels{
		Ingress: make(shila.PacketChannelPubChannel),
		Egress:	 make(shila.PacketChannelPubChannel),
	}

	endpointIssues := shila.EndpointIssuePubChannels{
		Ingress: make(shila.EndpointIssuePubChannel),
		Egress:  make(shila.EndpointIssuePubChannel),
	}

	// Create and setup the kernelSide side
	kernelSide := kernelSide.New(trafficChannelPubs, endpointIssues)
	if err = kernelSide.Setup(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to setup kernel side.").Error())
		return ErrorCode
	}
	log.Verbose.Println("Kernel side setup successfully.")
	defer kernelSide.CleanUp()

	// Create and setup the network side
	networkSide := networkSide.New(trafficChannelPubs, endpointIssues)
	if err = networkSide.Setup(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to setup network side.").Error())
		return ErrorCode
	}
	log.Verbose.Println("Network side setup successfully.")
	defer networkSide.CleanUp()

	// Setup the ingress working side
	workingSideIngress := workingSide.New(connection.NewMapping(kernelSide, networkSide, netflow.NewRouter()),
		trafficChannelPubs.Ingress, endpointIssues.Ingress, workingSide.Ingress)
	if err := workingSideIngress.Setup(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to setup ingress working side.").Error())
		return ErrorCode
	}
	defer workingSideIngress.CleanUp()

	// Setup the egress working side
	workingSideEgress := workingSide.New(connection.NewMapping(kernelSide, networkSide, netflow.NewRouter()),
		trafficChannelPubs.Egress, endpointIssues.Egress, workingSide.Egress)
	if err := workingSideEgress.Setup(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to setup egress working side.").Error())
		return ErrorCode
	}
	defer workingSideEgress.CleanUp()
	log.Verbose.Println("Working sides setup successfully.")

	log.Verbose.Println("Setup done, starting machinery..")

	if err = workingSideIngress.Start(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to start ingress working side.").Error())
		return ErrorCode
	}
	if err = workingSideEgress.Start(); err != nil {
		log.Error.Print(shila.PrependError(err, "Unable to start egress working side.").Error())
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

	log.Info.Println("Shila up and running.")

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

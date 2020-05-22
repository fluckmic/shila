package shila

import (
	"shila/config"
)

// Defines all the interfaces which the network endpoint generator has to implement as they are used
// by the manager of the network side.

type NetworkEndpointGenerator interface {
	NewClient(netConnId NetFlow, l EndpointLabel, c config.NetworkEndpoint) ClientNetworkEndpoint
	NewServer(netConnId NetFlow, l EndpointLabel, c config.NetworkEndpoint) ServerNetworkEndpoint
}

// Should be able to create a network address from a string.
type NetworkAddressGenerator interface {
	NewAddress(string) 								  NetworkAddress     	// Generates a new network address from a string
	NewLocalAddress(int) 						  	  NetworkAddress        // Generates a new local network address from a string
	GenerateRemoteContactingAddress(address NetworkAddress) NetworkAddress 	// Generates the remote contacting network address
}

type NetworkNetFlowGenerator interface {
	LocalContactingNetFlow() NetFlow
	GenerateRemoteContactingFlow(NetFlow) NetFlow
}

// Should be able to create a network path from a string.
type NetworkPathGenerator interface {
	NewPath(string) NetworkPath
	GetDefaultContactingPath(address NetworkAddress) NetworkPath
}

type ClientNetworkEndpoint interface {
	Endpoint
	SetupAndRun() 	(NetFlow, error)
}

type ServerNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
}

type NetworkAddress interface {
	String() string
}

type NetworkPath interface {
	String() string
}

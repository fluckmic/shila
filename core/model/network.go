package model

import (
	"shila/config"
)

// Mappings
type ServerEndpointMapping map[NetworkAddressKey] 			ServerNetworkEndpoint
type ClientEndpointMapping map[NetworkAddressAndPathKey]	ClientNetworkEndpoint

type NetworkEndpointGenerator interface {
	NewClient(netConnId NetworkConnectionIdentifier, l EndpointLabel, c config.NetworkEndpoint) ClientNetworkEndpoint
	NewServer(netConnId NetworkConnectionIdentifier, l EndpointLabel, c config.NetworkEndpoint) ServerNetworkEndpoint
}

// Should be able to create a network address from a string.
type NetworkAddressGenerator interface {
	NewAddress(string) 								  NetworkAddress 		// Generates a new network address from a string
	NewLocalAddress(string) 						  NetworkAddress
	GenerateContactingAddress(address NetworkAddress) NetworkAddress 		// Generates the contacting network address from (traffic) network address
}

// Should be able to create a network path from a string.
type NetworkPathGenerator interface {
	NewPath(string) NetworkPath
	GetDefaultContactingPath(address NetworkAddress) NetworkPath
}

type ClientNetworkEndpoint interface {
	Endpoint
	SetupAndRun() 	(NetworkConnectionIdentifier, error)
}

type ServerNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
	RegisterConnection(NetworkConnectionIdentifier) error
}

type NetworkAddress interface {
	String() string
}

type NetworkPath interface {
	String() string
}

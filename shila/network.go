package shila

import "shila/config"

type NetworkEndpointGenerator interface {
	NewClient(connectTo NetworkAddress, connectVia NetworkPath, l EndpointLabel, c config.NetworkEndpoint) ClientNetworkEndpoint
	NewServer(listenTo NetworkAddress, l EndpointLabel, c config.NetworkEndpoint) ServerNetworkEndpoint
}

// Should be able to create a network address from a string.
type NetworkAddressGenerator interface {
	NewAddress(string) 		NetworkAddress
	NewLocalAddress(string) NetworkAddress
}

// Should be able to create a network path from a string.
type NetworkPathGenerator interface {
	NewPath(string) NetworkPath
}

type ClientNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
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

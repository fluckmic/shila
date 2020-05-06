package shila

import (
	"shila/config"
)

// Keys
type Key_NetworkAddress_ 		string
type Key_NetworkAddressAndPath_ string

// Mappings
type ServerEndpointMapping map[Key_NetworkAddress_] ServerNetworkEndpoint
type ClientEndpointMapping map[Key_NetworkAddressAndPath_] ClientNetworkEndpoint

type NetworkEndpointGenerator interface {
	NewClient(connectTo NetworkAddress, connectVia NetworkPath, l EndpointLabel, c config.NetworkEndpoint) ClientNetworkEndpoint
	NewServer(listenTo NetworkAddress, l EndpointLabel, c config.NetworkEndpoint) ServerNetworkEndpoint
}

// Should be able to create a network address from a string.
type NetworkAddressGenerator interface {
	NewAddress(string) 		NetworkAddress
	NewLocalAddress(string) NetworkAddress
}

type NetworkKeyGenerator interface {
	GetAddressKey(address NetworkAddress) 						Key_NetworkAddress_
	GetAddressPathKey(address NetworkAddress, path NetworkPath) Key_NetworkAddressAndPath_
}

// Should be able to create a network path from a string.
type NetworkPathGenerator interface {
	NewPath(string) NetworkPath
	GetDefaultContactingPath(address NetworkAddress) NetworkPath
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

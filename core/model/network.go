package model

import (
	"shila/config"
)

// Keys
// (ipv4:port)
type IPAddressKey			  string
// (network-address)
type NetworkAddressKey		  string
// (network-address<>path)
type NetworkAddressAndPathKey string

type ServerNetworkEndpointAndConnectionCount struct {
	Endpoint ServerNetworkEndpoint
	ConnectionCount int
}

// Mappings
type ServerEndpointMapping map[NetworkAddressKey] ServerNetworkEndpointAndConnectionCount
type ClientEndpointMapping map[NetworkAddressAndPathKey]	ClientNetworkEndpoint

type NetworkEndpointGenerator interface {
	NewClient(connectTo NetworkAddress, connectVia NetworkPath, l EndpointLabel, c config.NetworkEndpoint) ClientNetworkEndpoint
	NewServer(listenTo NetworkAddress, l EndpointLabel, c config.NetworkEndpoint) ServerNetworkEndpoint
}

// Should be able to create a network address from a string.
type NetworkAddressGenerator interface {
	NewAddress(string) 								  NetworkAddress 		// Generates a new network address from a string
	NewLocalAddress(string) 						  NetworkAddress
	GenerateContactingAddress(address NetworkAddress) NetworkAddress 		// Generates the contacting network address from (traffic) network address
}

type NetworkKeyGenerator interface {
	GetAddressKey(address NetworkAddress) NetworkAddressKey
	GetAddressPathKey(address NetworkAddress, path NetworkPath) NetworkAddressAndPathKey
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

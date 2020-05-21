package shila

import (
	"shila/config"
)

// Mappings
type ServerNetworkEndpointMapping struct {
	ServerNetworkEndpoint
	IPConnectionMapping
}

type IPConnectionMapping 		map[IPFlowKey]	bool
type ServerEndpointMapping		map[NetworkAddressKey]			ServerNetworkEndpointMapping
type ClientEndpointMapping 		map[IPFlowKey]					ClientNetworkEndpoint

func (sb *ServerNetworkEndpointMapping) AddIPFlowKey(key IPFlowKey) {
	sb.IPConnectionMapping[key] = true
}

func (sb *ServerNetworkEndpointMapping) RemoveIPFlowKey(key IPFlowKey) {
	delete(sb.IPConnectionMapping, key)
}

func (sb *ServerNetworkEndpointMapping) Empty() bool {
	return len(sb.IPConnectionMapping) == 0
}

type NetworkEndpointGenerator interface {
	NewClient(netConnId NetFlow, l EndpointLabel, c config.NetworkEndpoint) ClientNetworkEndpoint
	NewServer(netConnId NetFlow, l EndpointLabel, c config.NetworkEndpoint) ServerNetworkEndpoint
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

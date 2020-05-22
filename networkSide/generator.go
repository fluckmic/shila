package networkSide

import (
	"shila/config"
	"shila/core/shila"
	"shila/networkSide/networkEndpoint"
)

// The generator of the network endpoint has to implement the following interfaces
var _ shila.NetworkEndpointGenerator 	= (*Generator)(nil)
var _ shila.NetworkAddressGenerator 	= (*Generator)(nil)
var _ shila.NetworkPathGenerator 		= (*Generator)(nil)
var _ shila.NetworkNetFlowGenerator		= (*Generator)(nil)

type Error string
func (e Error) Error() string {
	return string(e)
}

const defaultPath 			= ""
const defaultContactingPort = 9876

type Generator struct{}

func (g Generator) GetDefaultContactingPath(address shila.NetworkAddress) shila.NetworkPath {
	_ = address
	return g.NewPath("")
}

func (g Generator) NewClient(netConnId shila.NetFlow, label shila.EndpointLabel, config config.NetworkEndpoint) shila.ClientNetworkEndpoint {
	return networkEndpoint.NewClient(netConnId, label, config)
}

func (g Generator) NewServer(netConnId shila.NetFlow, label shila.EndpointLabel,
	config config.NetworkEndpoint) shila.ServerNetworkEndpoint {
	return networkEndpoint.NewServer(netConnId, label, config)
}

func (g Generator) NewAddress(address string) shila.NetworkAddress {
	return newAddress(address)
}

func (g Generator) NewLocalAddress(port int) shila.NetworkAddress {
	return newLocalNetworkAddress(port)
}

func (g Generator) NewEmptyAddress() shila.NetworkAddress {
	return newEmptyNetworkAddress()
}

func (g Generator) NewPath(path string) shila.NetworkPath {
	return newPath(path)
}

func (g Generator) GenerateRemoteContactingAddress(address shila.NetworkAddress) shila.NetworkAddress {
	return generateContactingAddress(address, defaultContactingPort)
}

func (g Generator) LocalContactingNetFlow() shila.NetFlow {
	return shila.NetFlow{Src: g.NewLocalAddress(defaultContactingPort)}
}

func (g Generator) GenerateRemoteContactingFlow(flow shila.NetFlow) shila.NetFlow {
	flow.Path = g.GetDefaultContactingPath(flow.Dst)
	flow.Dst  = g.GenerateRemoteContactingAddress(flow.Dst)
	return flow
}
package networkEndpoint

import (
	"shila/config"
	"shila/core/model"
)

var _ model.NetworkEndpointGenerator 	= (*Generator)(nil)
var _ model.NetworkAddressGenerator 	= (*Generator)(nil)
var _ model.NetworkPathGenerator 		= (*Generator)(nil)

type Error string
func (e Error) Error() string {
	return string(e)
}

const defaultPath 			= ""
const defaultContactingPort = 9876

type Base struct {
	label           model.EndpointLabel
	ingress			model.PacketChannel
	egress 			model.PacketChannel
	config 			config.NetworkEndpoint
}
type Generator struct{}

func (g Generator) GetDefaultContactingPath(address model.NetworkAddress) model.NetworkPath {
	_ = address
	return g.NewPath("")
}

func (g Generator) NewClient(netConnId model.NetworkConnectionIdentifier, label model.EndpointLabel, config config.NetworkEndpoint) model.ClientNetworkEndpoint {
	return newClient(netConnId, label, config)
}

func (g Generator) NewServer(netConnId model.NetworkConnectionIdentifier, label model.EndpointLabel,
	config config.NetworkEndpoint) model.ServerNetworkEndpoint {
	return newServer(netConnId, label, config)
}

func (g Generator) NewAddress(address string) model.NetworkAddress {
	return newAddress(address)
}

func (g Generator) NewLocalAddress(port string) model.NetworkAddress {
	return newLocalNetworkAddress(port)
}

func (g Generator) NewEmptyAddress() model.NetworkAddress {
	return newEmptyNetworkAddress()
}

func (g Generator) NewPath(path string) model.NetworkPath {
	return newPath(path)
}

func (g Generator) GenerateContactingAddress(address model.NetworkAddress) model.NetworkAddress {
	return generateContactingAddress(address, defaultContactingPort)
}
package networkEndpoint

import (
	"shila/config"
	"shila/core/shila"
)

var _ shila.NetworkEndpointGenerator = (*Generator)(nil)
var _ shila.NetworkAddressGenerator = (*Generator)(nil)
var _ shila.NetworkPathGenerator = (*Generator)(nil)

type Error string
func (e Error) Error() string {
	return string(e)
}

const defaultPath 			= ""
const defaultContactingPort = 9876

type Base struct {
	label   shila.EndpointLabel
	ingress shila.PacketChannel
	egress  shila.PacketChannel
	config  config.NetworkEndpoint
}
type Generator struct{}

func (g Generator) GetDefaultContactingPath(address shila.NetworkAddress) shila.NetworkPath {
	_ = address
	return g.NewPath("")
}

func (g Generator) NewClient(netConnId shila.NetFlow, label shila.EndpointLabel, config config.NetworkEndpoint) shila.ClientNetworkEndpoint {
	return newClient(netConnId, label, config)
}

func (g Generator) NewServer(netConnId shila.NetFlow, label shila.EndpointLabel,
	config config.NetworkEndpoint) shila.ServerNetworkEndpoint {
	return newServer(netConnId, label, config)
}

func (g Generator) NewAddress(address string) shila.NetworkAddress {
	return newAddress(address)
}

func (g Generator) NewLocalAddress(port string) shila.NetworkAddress {
	return newLocalNetworkAddress(port)
}

func (g Generator) NewEmptyAddress() shila.NetworkAddress {
	return newEmptyNetworkAddress()
}

func (g Generator) NewPath(path string) shila.NetworkPath {
	return newPath(path)
}

func (g Generator) GenerateContactingAddress(address shila.NetworkAddress) shila.NetworkAddress {
	return generateContactingAddress(address, defaultContactingPort)
}
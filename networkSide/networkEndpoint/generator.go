package networkEndpoint

import (
	"shila/config"
	"shila/core/model"
)

var _ model.NetworkEndpointGenerator 	= (*Generator)(nil)
var _ model.NetworkAddressGenerator 	= (*Generator)(nil)
var _ model.NetworkPathGenerator 		= (*Generator)(nil)
var _ model.NetworkKeyGenerator			= (*Generator)(nil)

type Error string
func (e Error) Error() string {
	return string(e)
}

type Base struct {
	label           model.EndpointLabel
	trafficChannels model.TrafficChannels
	config 			config.NetworkEndpoint
}
type Generator struct{}

func (g Generator) GetDefaultContactingPath(address model.NetworkAddress) model.NetworkPath {
	_ = address
	return g.NewPath("")
}

func (g Generator) GetAddressKey(address model.NetworkAddress) model.NetworkAddressKey {
	return model.NetworkAddressKey(address.String())
}

func (g Generator) GetAddressPathKey(address model.NetworkAddress, path model.NetworkPath) model.NetworkAddressAndPathKey {
	_ = path
	return model.NetworkAddressAndPathKey(address.String())
}

func (g Generator) NewClient(connectTo model.NetworkAddress, connectVia model.NetworkPath,
	label model.EndpointLabel, config config.NetworkEndpoint) model.ClientNetworkEndpoint {
	return newClient(connectTo, connectVia, label, config)
}

func (g Generator) NewServer(listenTo model.NetworkAddress, label model.EndpointLabel,
	config config.NetworkEndpoint) model.ServerNetworkEndpoint {
	return newServer(listenTo, label, config)
}

func (g Generator) NewAddress(address string) model.NetworkAddress {
	return newAddress(address)
}

func (g Generator) NewLocalAddress(port string) model.NetworkAddress {
	return newLocalNetworkAddress(port)
}

func (g Generator) NewPath(path string) model.NetworkPath {
	return newPath(path)
}


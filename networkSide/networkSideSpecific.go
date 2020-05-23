package networkSide

import (
	"net"
	"shila/config"
	"shila/core/shila"
	"shila/networkSide/network"
	"shila/networkSide/networkEndpoint"
	"strconv"
)

var _ shila.SpecificNetworkSideManager = (*SpecificManager)(nil)

const defaultPath 			= ""
const defaultContactingPort = 9876

type SpecificManager struct {
	config config.Config
}

func NewSpecificManager(config config.Config) SpecificManager {
	return SpecificManager{config: config}
}

func (specMng SpecificManager) NewClient(netConnId shila.NetFlow, label shila.EndpointLabel) shila.ClientNetworkEndpoint {
	return networkEndpoint.NewClient(netConnId, label, specMng.config.NetworkEndpoint)
}

func (specMng SpecificManager) NewServer(netConnId shila.NetFlow, label shila.EndpointLabel) shila.ServerNetworkEndpoint {
	return networkEndpoint.NewServer(netConnId, label, specMng.config.NetworkEndpoint)
}

func (specMng SpecificManager) RemoteContactingFlow(flow shila.NetFlow) shila.NetFlow {
	remContFlow := flow
	remContFlow.Path = specMng.getDefaultContactingPath(flow.Dst)
	remContFlow.Dst  = specMng.generateRemoteContactingAddress(flow.Dst)
	return remContFlow
}

func (specMng SpecificManager) LocalContactingNetFlow() shila.NetFlow {
	return shila.NetFlow{Src: network.AddressGenerator{}.NewLocal(strconv.Itoa(defaultContactingPort))}
}

func (specMng SpecificManager) generateRemoteContactingAddress(address shila.NetworkAddress) shila.NetworkAddress {
	addr := address.(*net.TCPAddr)
	addr.Port = defaultContactingPort
	return addr
}

func (specMng SpecificManager) getDefaultContactingPath(address shila.NetworkAddress) shila.NetworkPath {
	_ = address
	return network.PathGenerator{}.NewEmpty()
}

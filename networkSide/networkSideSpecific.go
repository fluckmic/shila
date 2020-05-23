package networkSide

import (
	"net"
	"shila/config"
	"shila/core/shila"
	"shila/networkSide/network"
	"shila/networkSide/networkEndpoint"
	"strconv"
)

// Generator functionalities are thought to be used outside of the
// backbone protocol specific implementations (suffix "Specific").
var _ shila.NetworkEndpointGenerator 	= (*SpecificManager)(nil)
var _ shila.NetworkNetFlowGenerator		= (*SpecificManager)(nil)

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
	flow.Path = specMng.getDefaultContactingPath(flow.Dst)
	flow.Dst  = specMng.generateRemoteContactingAddress(flow.Dst)
	return flow
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

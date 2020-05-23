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

func (specMng SpecificManager) NewClient(netConnId shila.NetFlow, label shila.EndpointLabel) shila.NetworkClientEndpoint {
	return networkEndpoint.NewClient(netConnId, label, specMng.config.NetworkEndpoint)
}

func (specMng SpecificManager) NewServer(netConnId shila.NetFlow, label shila.EndpointLabel) shila.NetworkServerEndpoint {
	return networkEndpoint.NewServer(netConnId, label, specMng.config.NetworkEndpoint)
}

func (specMng SpecificManager) RemoteContactingFlow(flow shila.NetFlow) shila.NetFlow {
	return shila.NetFlow{
		Src:  flow.Src,
		Path: specMng.getDefaultContactingPath(flow.Dst),
		Dst:  specMng.generateRemoteContactingAddress(flow.Dst),
	}
}

func (specMng SpecificManager) LocalContactingNetFlow() shila.NetFlow {
	return shila.NetFlow{
		Src: network.AddressGenerator{}.NewLocal(strconv.Itoa(defaultContactingPort)),
	}
}

func (specMng SpecificManager) generateRemoteContactingAddress(address shila.NetworkAddress) shila.NetworkAddress {
	addr := address.(*net.TCPAddr)
	return &net.TCPAddr{
		IP:   addr.IP,
		Port: defaultContactingPort,
		Zone: addr.Zone,
	}
}

func (specMng SpecificManager) getDefaultContactingPath(address shila.NetworkAddress) shila.NetworkPath {
	_ = address
	return network.PathGenerator{}.NewEmpty()
}

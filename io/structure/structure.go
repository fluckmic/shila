package structure

import (
	"fmt"
	"net"
	"shila/core/shila"
	"shila/networkSide/network"
	"strconv"
)

type IPAddressPortJSON struct {
	IP   string
	Port string
}
func (ipj IPAddressPortJSON) GetIPAddressPort() (net.TCPAddr, error) {

	IPv4 := net.ParseIP(ipj.IP)
	Port, err := strconv.Atoi(ipj.Port)
	if IPv4 == nil {
		return net.TCPAddr{}, shila.ThirdPartyError("Unable to parse port.")
	} else if err != nil {
		return net.TCPAddr{}, err
	}

	return net.TCPAddr{
		IP:   IPv4,
		Port: Port,
		Zone: "",
	}, nil
}

type NetworkPathJSON struct {
	Entries []NetworkPathEntryJSON
}
func (pj NetworkPathJSON) GetPath() (shila.NetworkPath, error) {
	return network.PathGenerator{}.New(""), nil
}
type NetworkPathEntryJSON struct {
	Entry interface{}
}

type NetworkAddressAndPathJSON struct {
	Address NetworkAddressJSON
	Path    NetworkPathJSON
}
func (nfj NetworkAddressAndPathJSON) GetNetworkAddressAndPath() (shila.NetworkAddress, shila.NetworkPath, error) {

	address, err := nfj.Address.GetNetworkAddress()
	if err != nil {
		return nil, nil, err
	}

	path, err := nfj.Path.GetPath()
	if err != nil {
		return nil, nil, err
	}

	return address, path, nil
}

type NetworkAddressJSON struct {
	IP   string
	Port string
}
func (naj NetworkAddressJSON) GetNetworkAddress() (shila.NetworkAddress, error) {
	return network.AddressGenerator{}.New(fmt.Sprint(naj.IP, ":", naj.Port))
}

type RoutingEntryJSON struct {
	Key  IPAddressPortJSON
	Flow NetworkAddressAndPathJSON
}

//
package structure

import (
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
	Elements []NetworkPathEntryJSON
}
func (pj NetworkPathJSON) GetPath() (shila.NetworkPath, error) {
	return network.PathGenerator{}.New("")
}
type NetworkPathEntryJSON struct {
	Element interface{}
}

type NetworkAddressAndPathJSON struct {
	Address string
	Path    NetworkPathJSON
}
func (json NetworkAddressAndPathJSON) GetNetworkAddressAndPath() (shila.NetworkAddress, shila.NetworkPath, error) {

	address, err := network.AddressGenerator{}.New(json.Address)
	if err != nil {
		return nil, nil, err
	}

	path, err := json.Path.GetPath()
	if err != nil {
		return nil, nil, err
	}

	return address, path, nil
}

type RoutingEntryJSON struct {
	Key  IPAddressPortJSON
	Flow NetworkAddressAndPathJSON
}

package networkEndpoint

import (
	"fmt"
	"net"
	"shila/core/model"
	"shila/log"
	"strconv"
)

var _ model.NetworkAddress = (*Address)(nil)

type Address struct {
	Addr net.TCPAddr
}

// <ip>:<port>
func newAddress(address string) model.NetworkAddress {

	if host, port, err := net.SplitHostPort(address); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new address from ", address, "."))
		return nil
	} else {
		IPv4 := net.ParseIP(host)
		Port, err := strconv.Atoi(port)
		if IPv4 != nil || err != nil {
			log.Error.Panic(fmt.Sprint("Unable to create new address from ", address, "."))
			return nil
		} else {
			return Address{Addr: net.TCPAddr{IP: IPv4, Port: Port}}
		}
	}

	return nil
}

// <port>
func newLocalNetworkAddress(port string) model.NetworkAddress {
	if Port, err := strconv.Atoi(port); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new local address from ", port, "."))
		return nil
	} else {
		return Address{Addr: net.TCPAddr{Port: Port}}
	}
	return nil
}

func (a Address) String() string {
	return a.Addr.String()
}


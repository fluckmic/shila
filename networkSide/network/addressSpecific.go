package network

import (
	"fmt"
	"net"
	"shila/core/shila"
	"shila/log"
	"strconv"
)

// Generator functionalities are thought to be used outside of the
// backbone protocol specific implementations (suffix "Specific").
var _ shila.NetworkAddressGenerator = (*AddressGenerator)(nil)
var _ shila.NetworkAddress 			= (*net.TCPAddr)(nil)

type AddressGenerator struct {}

// <ip>:<port>
func (g AddressGenerator) New(address string) shila.NetworkAddress {
	return newAddress(address)
}

func newAddress(addr string) shila.NetworkAddress {
	if host, port, err := net.SplitHostPort(addr); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new network address from {", addr, "}."))
		return nil
	} else {
		IPv4 := net.ParseIP(host)
		Port, err := strconv.Atoi(port)
		if IPv4 == nil || err != nil {
			log.Error.Panic(fmt.Sprint("Unable to create new network address from {", addr, "}."))
			return nil
		} else {
			return &net.TCPAddr{IP: IPv4, Port: Port}
		}
		return nil
	}
}

// <port>
func (g AddressGenerator) NewLocal(port string) shila.NetworkAddress {
		if Port, err := strconv.Atoi(port); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new local address from {", port, "}."))
		return nil
	} else {
		return &net.TCPAddr{Port: Port}
		}
	}

func (g AddressGenerator) NewEmpty() shila.NetworkAddress {
	return &net.TCPAddr{}
}


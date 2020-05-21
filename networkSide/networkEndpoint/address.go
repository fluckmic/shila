package networkEndpoint

import (
	"fmt"
	"net"
	"shila/core/shila"
	"shila/log"
	"strconv"
)

var _ shila.NetworkAddress = (*Address)(nil)

type Address struct {
	Addr net.TCPAddr
}

// <ip>:<port>
func newAddress(address string) shila.NetworkAddress {

	if host, port, err := net.SplitHostPort(address); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new network address from {", address, "}."))
		return nil
	} else {
		IPv4 := net.ParseIP(host)
		Port, err := strconv.Atoi(port)
		if IPv4 == nil || err != nil {
			log.Error.Panic(fmt.Sprint("Unable to create new network address from {", address, "}."))
			return nil
		} else {
			return Address{Addr: net.TCPAddr{IP: IPv4, Port: Port}}
		}
	}
	return nil
}

// <port>
func newLocalNetworkAddress(port string) shila.NetworkAddress {
	if Port, err := strconv.Atoi(port); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new local address from ", port, "."))
		return nil
	} else {
		return Address{Addr: net.TCPAddr{Port: Port}}
	}
	return nil
}

func generateContactingAddress(address shila.NetworkAddress, port int) shila.NetworkAddress {
	if host, _, err := net.SplitHostPort(address.String()); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to generate contacting address from {", address, "}."))
		return nil
	} else {
		if IPv4 := net.ParseIP(host); IPv4 == nil {
			log.Error.Panic(fmt.Sprint("Unable to generate contacting address from {", address, "}."))
			return nil
		} else {
			return Address{Addr: net.TCPAddr{IP: IPv4, Port: port}}
		}
	}
}

func newEmptyNetworkAddress() shila.NetworkAddress {
	return Address{Addr: net.TCPAddr{}}
}



func (a Address) String() string {
	return a.Addr.String()
}

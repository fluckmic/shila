//
package network

import (
	"fmt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/snet"
	"net"
	"shila/core/shila"
	"strconv"
)

// Generator functionalities are thought to be used outside of the
// backbone protocol specific implementations (suffix "Specific").
var _ shila.NetworkAddressGenerator = (*AddressGenerator)(nil)
var _ shila.NetworkAddress 			= (*snet.UDPAddr)(nil)

type AddressGenerator struct {}

func (g AddressGenerator) New(address string) (shila.NetworkAddress, error) {
	return snet.ParseUDPAddr(address)
}

func (g AddressGenerator) NewLocal(portStr string) (shila.NetworkAddress, error) {

	if port, err := strconv.Atoi(portStr); err != nil {
		return &snet.UDPAddr{}, shila.PrependError(err, fmt.Sprint("Cannot parse port ", portStr, "."))
	} else {
		return &snet.UDPAddr{
			IA:      addr.IA{},
			Path:    nil,
			NextHop: nil,
			Host:    &net.UDPAddr{Port: port},
		}, nil
	}
}

func (g AddressGenerator) NewEmpty() shila.NetworkAddress {
	return &snet.UDPAddr{}
}

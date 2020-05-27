package kernelEndpoint

import (
	"fmt"
	"net"
	"shila/core/shila"
	"shila/kernelSide/namespace"
)

type Identifier struct {
	number    uint
	namespace namespace.Namespace
	ip        net.IP
}

func NewIdentifier(number uint, namespace namespace.Namespace, ip net.IP) Identifier {
	return Identifier{number, namespace, ip}
}

func (id *Identifier) Name() string {
	return fmt.Sprint("tun", id.number)
}

func (id *Identifier) Number() uint {
	return id.number
}

func (id *Identifier) Namespace() string {
	if !id.namespace.NonEmpty {
		return ""
	} else {
		return id.namespace.Name
	}
}

func (id *Identifier) InDefaultNamespace() bool {
	return !id.namespace.NonEmpty
}

func (id *Identifier) IP() string {
	return id.ip.String()
}

// TODO: What is the best key to use?
func (id *Identifier) Key() shila.IPAddressKey {
	return shila.GetIPAddressKey(id.ip)
}

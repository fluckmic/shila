package kerep

import (
	"fmt"
	"net"
	"shila/helper"
)

type Identifier struct {
	number    uint
	namespace *helper.Namespace
	ip        net.IP
}

func NewIdentifier(number uint, namespace *helper.Namespace, ip net.IP) Identifier {
	return Identifier{number, namespace, ip}
}

func (id *Identifier) Name() string {
	return fmt.Sprint("tun", id.number)
}

func (id *Identifier) Number() uint {
	return id.number
}

func (id *Identifier) Namespace() string {
	if id.namespace == nil {
		return ""
	} else {
		return id.namespace.Name
	}
}

func (id *Identifier) InDefaultNamespace() bool {
	return id.namespace == nil
}

func (id *Identifier) IP() string {
	return id.ip.String()
}

// TODO: What is the best key to use?
func (id *Identifier) Key() string {
	return id.IP()
}

package kerep

import (
	"fmt"
	"net"
	"shila/helper"
)

type Identifier struct {
	number    uint
	namespace *helper.Namespace
	subnet    net.IPNet
}

func NewIdentifier(number uint, namespace *helper.Namespace, subnet net.IPNet) Identifier {
	return Identifier{number, namespace, subnet}
}

func (id *Identifier) Name() string {
	return fmt.Sprint("tun", id.number)
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

func (id *Identifier) Subnet() string {
	return id.subnet.String()
}

// TODO: What is the best key to use?
func (id *Identifier) Key() string {
	return fmt.Sprint(id.Namespace(), "-", id.Subnet())
}

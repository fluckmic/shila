package appep

import (
	"fmt"
	"net"
	"shila/appep/vif"
)

type Identifier struct {
	number    uint
	namespace *vif.Namespace
	subnet    net.IPNet
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

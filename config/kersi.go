package config

import (
	"net"
	"shila/helper"
)

var _ config = (*KernelSide)(nil)

type KernelSide struct {
	NEgressKerEp uint

	EgressNamespace  *helper.Namespace
	IngressNamespace *helper.Namespace

	EgressIP  net.IP
	IngressIP net.IP
}

func (k *KernelSide) InitDefault() error {

	k.NEgressKerEp = 3

	k.EgressNamespace = &helper.Namespace{Name: "shila-egress"}
	k.IngressNamespace = &helper.Namespace{Name: "shila-ingress"}

	k.EgressIP = net.IPv4(10, 0, 0, 1)
	k.IngressIP = net.IPv4(10, 7, 0, 9)

	return nil
}

package config

import (
	"net"
	"shila/helper"
)

var _ config = (*KernelSide)(nil)

type KernelSide struct {
	NEgressKerEp  uint
	NIngressKerEp uint

	EgressNamespace  *helper.Namespace
	IngressNamespace *helper.Namespace

	StartingEgressSubnet  net.IPNet
	StartingIngressSubnet net.IPNet
}

func (k *KernelSide) InitDefault() error {

	k.NEgressKerEp = 3
	k.NIngressKerEp = 1

	k.EgressNamespace = &helper.Namespace{Name: "shila-egress"}
	k.IngressNamespace = &helper.Namespace{Name: "shila-ingress"}

	k.StartingEgressSubnet = net.IPNet{
		IP:   net.IPv4(10, 0, 0, 1),
		Mask: net.CIDRMask(32, 32),
	}

	k.StartingIngressSubnet = net.IPNet{
		IP:   net.IPv4(10, 7, 0, 9),
		Mask: net.CIDRMask(32, 32),
	}

	return nil
}

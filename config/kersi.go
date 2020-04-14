package config

import (
	"net"
	"shila/kersi/kerep/vif"
)

var _ config = (*KernelSide)(nil)

type KernelSide struct {
	NEgressKerEp  uint
	NIngressKerEp uint

	EgressNamespace  *vif.Namespace
	IngressNamespace *vif.Namespace

	StartingEgressSubnet  net.IPNet
	StartingIngressSubnet net.IPNet
}

func (k *KernelSide) InitDefault() error {

	k.NEgressKerEp = 3
	k.NIngressKerEp = 1

	k.EgressNamespace = &vif.Namespace{Name: "shila-egress"}
	k.IngressNamespace = &vif.Namespace{Name: "shila-ingress"}

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

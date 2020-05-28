package kernelSide

import (
	"net"
	"shila/kernelSide/network"
)

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	NEgressKerEp     uint
	EgressNamespace  network.Namespace
	IngressNamespace network.Namespace
	EgressIP         net.IP
	IngressIP        net.IP
}

func hardCodedConfig() config {
	return config{
		NEgressKerEp:     3,
		EgressNamespace:  network.NewNamespace("shila-egress"),
		IngressNamespace: network.NewNamespace("shila-ingress"),
		EgressIP:         net.IPv4(10, 0, 0, 1),
		IngressIP:        net.IPv4(10, 7, 0, 9),
	}
}

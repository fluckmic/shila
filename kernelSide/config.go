package kernelSide

import (
	"net"
	"shila/kernelSide/namespace"
)

type Config struct {
	NEgressKerEp 		uint
	EgressNamespace  	namespace.Namespace
	IngressNamespace 	namespace.Namespace
	EgressIP  			net.IP
	IngressIP 			net.IP
}

func HardCodedConfig() Config {
	return Config{
		NEgressKerEp:     3,
		EgressNamespace:  namespace.NewNamespace("shila-egress"),
		IngressNamespace: namespace.NewNamespace("shila-ingress"),
		EgressIP:         net.IPv4(10, 0, 0, 1),
		IngressIP:        net.IPv4(10, 7, 0, 9),
	}
}

package kernelSide

import (
	"net"
	"shila/kernelSide/namespace"
)

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	NEgressKerEp 		uint
	EgressNamespace  	namespace.Namespace
	IngressNamespace 	namespace.Namespace
	EgressIP  			net.IP
	IngressIP 			net.IP
}

func hardCodedConfig() config {
	return config{
		NEgressKerEp:     3,
		EgressNamespace:  namespace.NewNamespace("shila-egress"),
		IngressNamespace: namespace.NewNamespace("shila-ingress"),
		EgressIP:         net.IPv4(10, 0, 0, 1),
		IngressIP:        net.IPv4(10, 7, 0, 9),
	}
}

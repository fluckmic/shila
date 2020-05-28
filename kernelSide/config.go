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
	NumberOfEgressInterfaces uint8					// Number of the egress virtual interfaces.
	EgressNamespace          network.Namespace		// The name of the egress namespace.
	IngressNamespace         network.Namespace		// The name of the ingress namespace.
	IngressIP                net.IP					// The IP of the ingress virtual interface.
}

func hardCodedConfig() config {
	return config{
		NumberOfEgressInterfaces: 3,
		EgressNamespace:          network.NewNamespace("shila-egress"),
		IngressNamespace:         network.NewNamespace("shila-ingress"),
		IngressIP:                net.IPv4(10, 7, 0, 9),
	}
}

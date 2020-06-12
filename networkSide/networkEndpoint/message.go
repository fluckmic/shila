package networkEndpoint

import (
	"net"
	"shila/core/shila"
)

type controlMessage struct {
	IPFlow          shila.IPFlow
	FlowKind        shila.FlowType
	LAddrContactEnd net.UDPAddr
}

type payloadMessage struct {
	Payload []byte
}
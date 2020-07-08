package networkEndpoint

import (
	"net"
	"shila/core/shila"
)

type controlMessage struct {
	IPFlow          shila.IPFlow
	LAddrContactEnd net.UDPAddr
	Payload 		[]byte
}

type payloadMessage struct {
	Payload []byte
}
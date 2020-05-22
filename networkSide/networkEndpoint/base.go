package networkEndpoint

import (
	"shila/config"
	"shila/core/shila"
)

type Base struct {
	label   shila.EndpointLabel
	ingress shila.PacketChannel
	egress  shila.PacketChannel
	config  config.NetworkEndpoint
}

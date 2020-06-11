package networkEndpoint

import (
	"shila/core/shila"
)

type Base struct {
	Role    shila.EndpointRole
	Ingress shila.PacketChannel
	Egress  shila.PacketChannel
	State   shila.EntityState
	Issues  shila.EndpointIssuePubChannel
}

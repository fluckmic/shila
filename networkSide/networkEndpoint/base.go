package networkEndpoint

import (
	"shila/core/shila"
)

type NetworkEndpointBase struct {
	role    shila.EndpointRole
	ingress shila.PacketChannel
	egress  shila.PacketChannel
	state   shila.EntityState
	issues  shila.EndpointIssuePubChannel
}

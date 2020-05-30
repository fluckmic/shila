package networkEndpoint

import (
	"shila/core/shila"
)

type Base struct {
	label   		shila.EndpointLabel
	ingress 		shila.PacketChannel
	egress  		shila.PacketChannel
	state   		shila.EntityState
	endpointIssues 	shila.EndpointIssuePubChannel
}

package networkSide

import (
	"shila/shila"
)

type Manager struct {
	Endpoints map[_IPv4_] *shila.NetworkEndpoint
}

type _IPv4_ string

type Error string

func (e Error) Error() string {
	return string(e)
}

func (m *Manager) EstablishNewServerEndpoint(addr shila.NetworkAddress) (shila.TrafficChannels, error) {
	return shila.TrafficChannels{}, nil
}

func (m *Manager) EstablishNewContactingClientEndpoint(addr shila.NetworkAddress) (shila.TrafficChannels, error) {
	/*
		// TODO: Config!
		egress  :=  make(shila.PacketChannel, 10)
		ingress :=  make(shila.PacketChannel, 10)
		// TODO: "default" path here, since this is the connection used for the contacting
		c.channels.Contacting.Endpoint = networkEndpoint.Generator{}.NewClient(c.header.Dst, c.header.Path, shila.ContactingNetworkEndpoint,
			shila.TrafficChannels{Ingress: c.channels.Contacting.Channels.Egress, Egress: c.channels.Contacting.Channels.Ingress})

		if err := c.channels.Contacting.Endpoint.SetupAndRun(); err != nil {
			return Error(fmt.Sprint("Failed to establish contacting channel - ", err.Error()))
		}
	*/
	return shila.TrafficChannels{}, nil
}

func (m *Manager) EstablishNewTrafficClientEndpoint(addr shila.NetworkAddress, path shila.NetworkPath) (shila.TrafficChannels, error) {

	// Close the contacting client endpoint as soon as the traffic client endpoint is established.

	return shila.TrafficChannels{}, nil
}
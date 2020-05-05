package config


var _ config = (*NetworkEndpoint)(nil)

type NetworkEndpoint struct {

	// kerep <-> core (in shila.packet)
	SizeIngressBuff int
	SizeEgressBuff  int

}

func (n *NetworkEndpoint) InitDefault() error {

	n.SizeIngressBuff = 10
	n.SizeEgressBuff = 10

	return nil
}

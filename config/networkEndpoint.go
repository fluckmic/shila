package config


var _ config = (*NetworkEndpoint)(nil)

type NetworkEndpoint struct {

	// kerep <-> core (in shila.packet)
	SizeIngressBuff int
	SizeEgressBuff  int

	// kernel <-> kerep (in Byte)
	SizeReadBuffer int
	BatchSizeRead  int

	SizeHoldingArea int

}

func (n *NetworkEndpoint) InitDefault() error {

	n.SizeIngressBuff = 10
	n.SizeEgressBuff = 10

	n.SizeReadBuffer = 1500
	n.BatchSizeRead = 30

	n.SizeHoldingArea = 100

	return nil
}

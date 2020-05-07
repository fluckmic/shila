package config

var _ config = (*KernelEndpoint)(nil)

type KernelEndpoint struct {

	// kerep <-> core (in shila.packet)
	SizeIngressBuff int
	SizeEgressBuff  int

	// kernel <-> kerep (in Byte)
	SizeReadBuffer int
	BatchSizeRead  int

	MaxNVifReader int
	MaxNVifWriter int
}

func (k *KernelEndpoint) InitDefault() error {

	k.SizeIngressBuff = 10
	k.SizeEgressBuff = 10

	k.SizeReadBuffer = 1500
	k.BatchSizeRead = 30

	k.MaxNVifReader = 1
	k.MaxNVifWriter = 1

	return nil
}

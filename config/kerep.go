package config

var _ config = (*KernelEndpoint)(nil)

type KernelEndpoint struct {
	SizeIngressBuff uint
	SizeEngressBuff uint

	MaxNVifReader uint
	MaxNVifWriter uint
}

func (k *KernelEndpoint) InitDefault() error {

	k.SizeIngressBuff = 10
	k.SizeEngressBuff = 10

	k.MaxNVifReader = 3
	k.MaxNVifWriter = 3

	return nil
}

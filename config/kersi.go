package config

var _ config = (*KernelSide)(nil)

type KernelSide struct {
}

func (k *KernelSide) InitDefault() error {
	return nil
}

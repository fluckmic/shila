package config

var _ config = (*Appep)(nil)

type Appep struct {
	SizeIngressBuff uint
	SizeEngressBuff uint

	MaxNVifReader uint
	MaxNVifWriter uint
}

func (c *Appep) InitDefault() error {

	c.SizeIngressBuff = 10
	c.SizeEngressBuff = 10

	c.MaxNVifReader = 3
	c.MaxNVifWriter = 3

	return nil
}

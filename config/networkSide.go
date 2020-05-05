package config

var _ config = (*NetworkSide)(nil)

type NetworkSide struct {

}

func (n *NetworkSide) InitDefault() error {
	return nil
}

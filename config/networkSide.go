package config

var _ config = (*NetworkSide)(nil)

type NetworkSide struct {
	ContactingServerPort	string
}

func (n *NetworkSide) InitDefault() error {
	n.ContactingServerPort = "9876"
	return nil
}

package config

var _ config = (*NetworkSide)(nil)

type NetworkSide struct {
	ContactingServerPort							string
	WaitingTimeUntilTrafficConnectionEstablishment 	int
}

func (n *NetworkSide) InitDefault() error {
	n.ContactingServerPort = "9876"
	n.WaitingTimeUntilTrafficConnectionEstablishment = 2
	return nil
}

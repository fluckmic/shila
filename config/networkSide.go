package config

var _ config = (*NetworkSide)(nil)

type NetworkSide struct {
	ContactingServerPort                string
	WaitingTimeTrafficConnEstablishment int
}

func (n *NetworkSide) InitDefault() error {
	n.ContactingServerPort 				  = "9876"
	n.WaitingTimeTrafficConnEstablishment = 2
	return nil
}

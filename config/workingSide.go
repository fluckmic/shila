package config

var _ config = (*WorkingSide)(nil)

type WorkingSide struct {

	NumberOfKernelEndpointIngressHandler 	int
	NumberOfKernelEndpointEgressHandler  	int
	NumberOfNetworkEndpointIngressHandler 	int
	NumberOfNetworkEndpointEgressHandler 	int

}

func (ws *WorkingSide) InitDefault() error {

	ws.NumberOfKernelEndpointEgressHandler 	 = 3
	ws.NumberOfKernelEndpointIngressHandler  = 3
	ws.NumberOfNetworkEndpointIngressHandler = 3
	ws.NumberOfNetworkEndpointEgressHandler  = 3

	return nil
}

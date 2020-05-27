package networkEndpoint

var ConfigLoaded bool
var Config config

func init() {
	Config 		 = hardCodedConfig()
	ConfigLoaded = true
}

type config struct {
	// kerep <-> core (in shila.packet)
	SizeIngressBuff int
	SizeEgressBuff  int
	// kernel <-> kerep (in Byte)
	SizeReadBuffer	int
	BatchSizeRead  	int
	SizeHoldingArea int
}

func hardCodedConfig() config {
	return config{
		SizeIngressBuff: 10,
		SizeEgressBuff:  10,
		SizeReadBuffer:  1500,
		BatchSizeRead:   30,
		SizeHoldingArea: 100,
	}
}


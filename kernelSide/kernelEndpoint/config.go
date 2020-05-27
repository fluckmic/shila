package kernelEndpoint

type Config struct {
	// kerep <-> core (in shila.packet)
	SizeIngressBuff int
	SizeEgressBuff  int
	// kernel <-> kerep (in Byte)
	SizeReadBuffer 	int
	BatchSizeRead  	int
	MaxNVifReader 	int
	MaxNVifWriter 	int
}

func HardCodedConfig() Config {
	return Config{
		SizeIngressBuff:	10,
		SizeEgressBuff:		10,
		SizeReadBuffer:		1500,
		BatchSizeRead:		30,
		MaxNVifReader:		1,
		MaxNVifWriter:		1,
	}
}
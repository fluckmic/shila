package shila

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	InitialTTL int 		// Number of sending retries for a packet
}

func hardCodedConfig() config {
	return config{
		InitialTTL: 5,
	}
}
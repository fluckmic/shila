//
package workingSide

var Config config

func init() {
	Config = hardCodedConfig()
}

type config struct {
	NumberOfWorkerPerChannel int	// Number of worker per packet channel.
}

func hardCodedConfig() config {
	return config{
		NumberOfWorkerPerChannel: 1,
	}
}
package workingSide

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	NumberOfWorkerPerChannel int
}

func hardCodedConfig() config {
	return config{
		NumberOfWorkerPerChannel: 1,
	}
}
package workingSide

var ConfigLoaded bool
var Config config

func init() {
	Config 		 = hardCodedConfig()
	ConfigLoaded = true
}

type config struct {
	NumberOfWorkerPerChannel int
}

func hardCodedConfig() config {
	return config{
		NumberOfWorkerPerChannel: 1,
	}
}
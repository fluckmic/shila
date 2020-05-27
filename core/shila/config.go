package shila

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {}

func hardCodedConfig() config {
	return config{}
}
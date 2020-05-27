package log

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	PrintVerbose 		bool
}

func hardCodedConfig() config {
	return config{
		PrintVerbose:     false,
	}
}

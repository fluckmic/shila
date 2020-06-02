//
package log

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	PrintVerbose 		bool	// Print verbose messages.
}

func hardCodedConfig() config {
	return config{
		PrintVerbose:     true,
	}
}

//
package netflow

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	Path string		// Path from where to load the routing entries inserted at startup.
}

func hardCodedConfig() config {
	return config{
		Path : "routing.json",
	}
}

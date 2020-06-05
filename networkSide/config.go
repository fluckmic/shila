//
package networkSide

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	ContactingServerPort                int				// Default port on which shila is listening for incoming contacting connections.
}

func hardCodedConfig() config {
	return config{
		ContactingServerPort: 					9876,
	}
}

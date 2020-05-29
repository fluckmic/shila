//
package networkSide

import "time"

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	ContactingServerPort                int				// Default port on which shila is listening for incoming contacting connections.
	WaitingTimeTrafficConnEstablishment time.Duration	// Minimal waiting time before a connection establishment with the traffic server endpoint is attempted.
}

func hardCodedConfig() config {
	return config{
		ContactingServerPort: 					9876,
		WaitingTimeTrafficConnEstablishment: 	time.Second * 2,
	}
}

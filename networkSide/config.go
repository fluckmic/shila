package networkSide

var ConfigLoaded bool
var Config config

func init() {
	Config 		 = hardCodedConfig()
	ConfigLoaded = true
}

type config struct {
	ContactingServerPort                int
	WaitingTimeTrafficConnEstablishment int
}

func hardCodedConfig() config {
	return config{
		ContactingServerPort: 9876,
		WaitingTimeTrafficConnEstablishment: 2,
	}
}

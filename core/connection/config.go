package connection

import (
	"time"
)

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	VacuumInterval 	 time.Duration
	MaxTimeUntouched time.Duration
}

func hardCodedConfig() config {
	return config{
		VacuumInterval: 	time.Second,
		MaxTimeUntouched: 	time.Second * 20,
	}
}

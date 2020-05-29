//
package connection

import (
	"time"
)

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	VacuumInterval 	 time.Duration		// Minimal amount of time between two vacuum processes.
	MaxTimeUntouched time.Duration		// Maximal amount of time a connection can stay untouched before it is closed.
}

func hardCodedConfig() config {
	return config{
		VacuumInterval: 	time.Second,
		MaxTimeUntouched: 	time.Second * 20,
	}
}

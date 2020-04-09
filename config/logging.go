package config

var _ config = (*Logging) (nil)

type Logging struct {

	// Config for log written to the log file
	DebugToFile 	bool
	InfoToFile  	bool
	FlagsFileLogger int

	// Config for the log written to the stdout
	DebugToStdout     bool
	InfoToStdout  	  bool
	FlagsStdoutLogger int
}

func (c *Logging) InitDefault() error {

	c.DebugToFile = false
	c.InfoToFile  = false

	c.DebugToStdout = true
	c.InfoToStdout  = true

	return nil
}
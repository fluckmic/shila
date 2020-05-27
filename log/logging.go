package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	Error   log.Logger
	Info    log.Logger
	Verbose log.Logger

	initialized bool
)

func Init() {

	Error.SetOutput(os.Stderr)
	Error.SetPrefix("ERROR: ")
	Error.SetFlags(log.Lshortfile)

	Info.SetOutput(os.Stdout)
	Info.SetPrefix("INFO: ")
	Info.SetFlags(log.Lshortfile)

	if !Config.PrintVerbose {
		Verbose.SetOutput(ioutil.Discard)
	} else {
		Verbose.SetOutput(os.Stdout)
		Verbose.SetPrefix("VERBOSE: ")
		Verbose.SetFlags(log.Lshortfile)
	}

	initialized = true
}

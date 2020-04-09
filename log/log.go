// TODO: Add a description.
package log

import (
	"log"
)

type Logger struct {
	log.Logger
}

func (l Logger) Debugln(v ...interface{}) {
		log.Println("DEBUG", v)
}

func (l Logger) Debugf(format string, v ...interface{}) {
		panic("TODO!") // TODO!
}

func (l Logger) Infoln(v ...interface{}) {
	log.Println("INFO", v)
}

func (l Logger) Infof(format string, v ...interface{}) {
		panic("TODO!") // TODO!
}
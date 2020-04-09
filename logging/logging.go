// TODO: Add a description.
package logging

import (
	"os"
	"shila/config"
	"shila/log"
)

type Logger struct {
	fileLogger *log.Logger
	stdLogger  *log.Logger
	config 	   config.Logging
}

func New(c config.Logging) *Logger {

	newLogger := &Logger{nil, nil, c}

	if c.DebugToFile || c.InfoToFile {
		panic("TODO!") // TODO!
	}
	if c.DebugToStdout || c.InfoToStdout {
		newLogger.stdLogger = new(log.Logger)
		newLogger.stdLogger.SetOutput(os.Stdout)
		newLogger.stdLogger.SetFlags(c.FlagsStdoutLogger)
	}

	return newLogger
}

func (l Logger) Debugln(v ...interface{}) {
	if l.config.DebugToStdout {
		l.stdLogger.Debugln(v)
	}
	if l.config.DebugToFile {
		l.fileLogger.Debugln(v)
	}
}

func (l Logger) Debugf(format string, v ...interface{}) {
	if l.config.DebugToStdout {
		panic("TODO!") // TODO!
	}
	if l.config.DebugToFile {
		panic("TODO!") // TODO!
	}
}

func (l Logger) Infoln(v ...interface{}) {
	if l.config.DebugToStdout {
		l.stdLogger.Infoln(v)
	}
	if l.config.DebugToFile {
		l.fileLogger.Infoln(v)
	}
}

func (l Logger) Infof(format string, v ...interface{}) {
	if l.config.InfoToStdout {
		panic("TODO!") // TODO!
	}
	if l.config.InfoToFile {
		panic("TODO!") // TODO!
	}
}
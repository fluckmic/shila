// TODO: Add a description.
package logging

import (
	"os"
	"shila/config"
	"shila/log"
)

type Error string
func (e Error) Error() string {
	return string(e)
}

type Logger struct {
	fileLogger *log.Logger
	stdLogger  *log.Logger
	config 	   config.Logging
}

func New(c config.Logging) (logger *Logger, err error) {

	defer func() {
		if e := recover(); e != nil {
			err = e.(Error)
			logger = nil
		}
	}()

	logger = &Logger{nil, nil, c}

	if c.DebugToFile || c.InfoToFile {
		panic("TODO!") // TODO!
	}
	if c.DebugToStdout || c.InfoToStdout {
		logger.stdLogger = new(log.Logger)
		logger.stdLogger.SetOutput(os.Stdout)
		logger.stdLogger.SetFlags(c.FlagsStdoutLogger)
	}
	return
}

func (l Logger) Debugln(v ...interface{}) {
	if l.config.DebugToStdout {
		l.stdLogger.Debugln(v...)
	}
	if l.config.DebugToFile {
		l.fileLogger.Debugln(v...)
	}
}

func (l Logger) Debugf(format string, v ...interface{}) {
	if l.config.DebugToStdout {
		l.stdLogger.Debugf(format, v...)
	}
	if l.config.DebugToFile {
		l.fileLogger.Debugf(format, v...)
	}
}

func (l Logger) Infoln(v ...interface{}) {

	if l.config.InfoToStdout {
		l.stdLogger.Infoln(v...)
	}
	if l.config.InfoToFile {
		l.fileLogger.Infoln(v...)
	}
}

func (l Logger) Infof(format string, v ...interface{}) {

	if l.config.InfoToStdout {
		l.stdLogger.Infof(format, v...)
	}
	if l.config.InfoToFile {
		l.fileLogger.Infof(format, v...)
	}
}
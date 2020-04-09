// TODO: Add a description.
package log

import (
	"fmt"
	"log"
)

type Logger struct {
	log.Logger
}

func (l Logger) Debugln(v ...interface{}) {
	v = append([]interface{}{"DEBUG"}, v...)
	log.Println(v...)}

func (l Logger) Debugf(format string, v ...interface{}) {
		log.Printf(fmt.Sprint("DEBUG ", format), v...)
}

func (l Logger) Infoln(v ...interface{}) {
	v = append([]interface{}{"INFO "}, v...)
	log.Println(v...)
}

func (l Logger) Infof(format string, v ...interface{}) {
		log.Printf(fmt.Sprint("INFO  ", format), v...)
}
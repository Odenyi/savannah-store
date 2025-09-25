package logger

import "log"

func Error(format string, v ...interface{}) {

	log.Printf(format,v...)

}

func Info(format string, v ...interface{}) {

	log.Printf(format,v...)
}
package logger

import (
	"log"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
)

var currentLevel = LevelInfo

func SetLevel(level string) {
	currentLevel = level
}

func Debug(format string, v ...interface{}) {
	if currentLevel == LevelDebug {
		log.Printf(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Fatal(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

package logger

import (
	"fmt"
	"log"
	"os"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var currentLevel = INFO

// Sets the minimum log level
func SetLevel(level Level) {
	currentLevel = level
}

// Logs debug messages
func Debug(format string, args ...interface{}) {
	if currentLevel <= DEBUG {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Logs info messages
func Info(format string, args ...interface{}) {
	if currentLevel <= INFO {
		log.Printf("[INFO] "+format, args...)
	}
}

// Logs warning messages
func Warn(format string, args ...interface{}) {
	if currentLevel <= WARN {
		log.Printf("[WARN] "+format, args...)
	}
}

// Logs error messages
func Error(format string, args ...interface{}) {
	if currentLevel <= ERROR {
		log.Printf("[ERROR] "+format, args...)
	}
}

// Logs error messages and exits
func Fatal(format string, args ...interface{}) {
	log.Printf("[FATAL] "+format, args...)
	os.Exit(1)
}

// Returns a formatted error
func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

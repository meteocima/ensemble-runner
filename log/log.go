package log

import (
	"fmt"
	"os"
	"time"
)

// Level is a type that represents
// the importance level of a log message
type Level int

const (
	// LevelError identify error messages
	LevelError Level = iota
	// LevelWarning identify Warning messages
	LevelWarning
	// LevelInfo identify Info messages
	LevelInfo
	// LevelDebug identify Debug messages
	LevelDebug
)

func (ll Level) String() string {
	switch ll {
	case LevelError:
		return "ERROR"
	case LevelWarning:
		return "WARNING"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	default:
		return "WRONGLEVEL"
	}
}

var level Level = LevelDebug

func write(msgLevel Level, msgText string, args []interface{}) {
	if msgLevel > level {
		return
	}
	dt := time.Now().Format(time.Stamp)
	fmt.Fprintf(os.Stderr, dt+" - "+msgLevel.String()+": "+msgText+"\n", args...)
}

// SetLevel set the maximum
// level a message must have to be
// logged.
func SetLevel(value Level) {
	level = value
}

// Debug prints a log string if
// the configured log level is
// equal or great than levelDebug
func Debug(msg string, args ...interface{}) {
	write(LevelDebug, msg, args)
}

// Info prints a log string if
// the configured log level is
// equal or great than levelInfo
func Info(msg string, args ...interface{}) {
	write(LevelInfo, msg, args)
}

// Warning prints a log string if
// the configured log level is
// equal or great than levelWarning
func Warning(msg string, args ...interface{}) {
	write(LevelWarning, msg, args)
}

// Error prints a log string if
// the configured log level is
// equal or great than levelError
func Error(msg string, args ...interface{}) {
	write(LevelError, msg, args)
}

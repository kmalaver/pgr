package pgr

import (
	"fmt"
	"log"
)

type LogLevel int

const (
	LogLevelNone LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelNone:
		return "none"
	default:
		return fmt.Sprintf("invalid level %d", ll)
	}
}

type kvs map[string]string

type Logger interface {
	Log(level LogLevel, data kvs)
}

type defLogger struct {
	level LogLevel
}

func (l *defLogger) Log(level LogLevel, data kvs) {
	if level >= l.level {
		log.Printf("[%s] %v", level.String(), data)
	}
}

package logger

import "context"

// Logger defines the logging interface
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields ...Field)
	
	// Info logs an info message
	Info(msg string, fields ...Field)
	
	// Warn logs a warning message
	Warn(msg string, fields ...Field)
	
	// Error logs an error message
	Error(msg string, fields ...Field)
	
	// WithFields returns a logger with additional fields
	WithFields(fields ...Field) Logger
	
	// WithContext returns a logger with context
	WithContext(ctx context.Context) Logger
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}
// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

// Option is a function type for configuring the Environment during initialization.
type Option func(*Environment)

// Logger is a function type for logging messages with different log levels.
type Logger func(level LogLevel, message string)

// LogLevel represents the severity level of a log message.
type LogLevel uint8

const (
	// LogLevelTrace represents the trace log level.
	LogLevelTrace LogLevel = iota
	// LogLevelDebug represents the debug log level.
	LogLevelDebug
	// LogLevelInfo represents the info log level.
	LogLevelInfo
	// LogLevelError represents the error log level.
	LogLevelError
	// LogLevelFatal represents the fatal log level.
	LogLevelFatal
)

// String converts a LogLevel value to a string.
func (l LogLevel) String() string {
	switch l {
	case LogLevelTrace:
		return "TRACE"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	}
	return "INFO"
}

// WithLogger is an Option function that sets the logger for the Environment.
func WithLogger(logger Logger) Option {
	return func(b *Environment) {
		b.Logger = logger
	}
}

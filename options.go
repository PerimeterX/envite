package fengshui

type Option func(*Blueprint)

type Logger func(level LogLevel, message string)

type LogLevel uint8

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelError
	LogLevelFatal
)

func WithLogger(logger Logger) Option {
	return func(b *Blueprint) {
		b.logger = logger
	}
}

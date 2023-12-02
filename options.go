package envite

type Option func(*Environment)

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
	return func(b *Environment) {
		b.Logger = logger
	}
}

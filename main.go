package alog

import "github.com/alley9040/ali-log/domain"

type LogLevel = domain.LogLevel
type LogField = domain.LogField
type LogConfig = domain.LogConfig
type Log = domain.Log

const (
	LogLevelDebug = domain.LogLevelDebug
	LogLevelInfo  = domain.LogLevelInfo
	LogLevelWarn  = domain.LogLevelWarn
	LogLevelError = domain.LogLevelError
	LogLevelFatal = domain.LogLevelFatal
	LogLevelPanic = domain.LogLevelPanic
)

func NewLogger(cfg *LogConfig) Log {
	return domain.NewLogger(cfg)
}

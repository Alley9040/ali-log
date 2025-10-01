package domain

type Log interface {
	Debug(msg string, fields ...LogField)
	Info(msg string, fields ...LogField)
	Warn(msg string, fields ...LogField)
	Error(msg string, fields ...LogField)
	Fatal(msg string, fields ...LogField)
	Panic(msg string, fields ...LogField)
	Printf(format string, args ...interface{})
	Close() error
}

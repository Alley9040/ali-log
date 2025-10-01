package domain

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota - 1 // -1
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelPanic
)

// String 返回日志级别的小写字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	case LogLevelPanic:
		return "panic"
	default:
		return "unknown"
	}
}

// ParseLogLevel 将字符串解析为 LogLevel（不区分大小写）
func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LogLevelDebug, nil
	case "info":
		return LogLevelInfo, nil
	case "warn", "warning":
		return LogLevelWarn, nil
	case "error":
		return LogLevelError, nil
	case "fatal":
		return LogLevelFatal, nil
	case "panic":
		return LogLevelPanic, nil
	default:
		return 0, fmt.Errorf("unknown log level: %s", s)
	}
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口，便于 mapstructure 使用
func (l *LogLevel) UnmarshalText(text []byte) error {
	if l == nil {
		return fmt.Errorf("nil LogLevel receiver")
	}
	lvl, err := ParseLogLevel(string(text))
	if err != nil {
		return err
	}
	*l = lvl
	return nil
}

type LogField zap.Field

func Error(err error) LogField {
	return LogField(zap.Error(err))
}

func String(key string, val string) LogField {
	return LogField(zap.String(key, val))
}

// 基本类型转换
func Int(key string, val int) LogField {
	return LogField(zap.Int(key, val))
}

func Int8(key string, val int8) LogField {
	return LogField(zap.Int8(key, val))
}

func Int16(key string, val int16) LogField {
	return LogField(zap.Int16(key, val))
}

func Int32(key string, val int32) LogField {
	return LogField(zap.Int32(key, val))
}

func Int64(key string, val int64) LogField {
	return LogField(zap.Int64(key, val))
}

func Uint(key string, val uint) LogField {
	return LogField(zap.Uint(key, val))
}

func Uint8(key string, val uint8) LogField {
	return LogField(zap.Uint8(key, val))
}

func Uint16(key string, val uint16) LogField {
	return LogField(zap.Uint16(key, val))
}

func Uint32(key string, val uint32) LogField {
	return LogField(zap.Uint32(key, val))
}

func Uint64(key string, val uint64) LogField {
	return LogField(zap.Uint64(key, val))
}

func Float32(key string, val float32) LogField {
	return LogField(zap.Float32(key, val))
}

func Float64(key string, val float64) LogField {
	return LogField(zap.Float64(key, val))
}

func Bool(key string, val bool) LogField {
	return LogField(zap.Bool(key, val))
}

func Complex64(key string, val complex64) LogField {
	return LogField(zap.Complex64(key, val))
}

func Complex128(key string, val complex128) LogField {
	return LogField(zap.Complex128(key, val))
}

// 时间类型转换
func Time(key string, val time.Time) LogField {
	return LogField(zap.Time(key, val))
}

func Duration(key string, val time.Duration) LogField {
	return LogField(zap.Duration(key, val))
}

// 接口类型转换
func Any(key string, val interface{}) LogField {
	return LogField(zap.Any(key, val))
}

func Binary(key string, val []byte) LogField {
	return LogField(zap.Binary(key, val))
}

func ByteString(key string, val []byte) LogField {
	return LogField(zap.ByteString(key, val))
}

// 切片类型转换
func Strings(key string, val []string) LogField {
	return LogField(zap.Strings(key, val))
}

func Ints(key string, val []int) LogField {
	return LogField(zap.Ints(key, val))
}

func Int64s(key string, val []int64) LogField {
	return LogField(zap.Int64s(key, val))
}

func Uints(key string, val []uint) LogField {
	return LogField(zap.Uints(key, val))
}

func Uint64s(key string, val []uint64) LogField {
	return LogField(zap.Uint64s(key, val))
}

func Float64s(key string, val []float64) LogField {
	return LogField(zap.Float64s(key, val))
}

func Bools(key string, val []bool) LogField {
	return LogField(zap.Bools(key, val))
}

func Times(key string, val []time.Time) LogField {
	return LogField(zap.Times(key, val))
}

func Durations(key string, val []time.Duration) LogField {
	return LogField(zap.Durations(key, val))
}

// 其他切片类型转换
func Uintptrs(key string, val []uintptr) LogField {
	return LogField(zap.Uintptrs(key, val))
}

func Complex128s(key string, val []complex128) LogField {
	return LogField(zap.Complex128s(key, val))
}

func Complex64s(key string, val []complex64) LogField {
	return LogField(zap.Complex64s(key, val))
}

func Float32s(key string, val []float32) LogField {
	return LogField(zap.Float32s(key, val))
}

func Errors(key string, val []error) LogField {
	return LogField(zap.Errors(key, val))
}

// 其他类型转换
func Array(key string, val zapcore.ArrayMarshaler) LogField {
	return LogField(zap.Array(key, val))
}

// 其他类型转换
func Uint8s(key string, val []uint8) LogField {
	return LogField(zap.Uint8s(key, val))
}

func Uint16s(key string, val []uint16) LogField {
	return LogField(zap.Uint16s(key, val))
}

func Uint32s(key string, val []uint32) LogField {
	return LogField(zap.Uint32s(key, val))
}

func Int8s(key string, val []int8) LogField {
	return LogField(zap.Int8s(key, val))
}

func Int16s(key string, val []int16) LogField {
	return LogField(zap.Int16s(key, val))
}

func Int32s(key string, val []int32) LogField {
	return LogField(zap.Int32s(key, val))
}

// 其他类型转换
func Uintp(key string, val *uint) LogField {
	return LogField(zap.Uintp(key, val))
}

func Uint8p(key string, val *uint8) LogField {
	return LogField(zap.Uint8p(key, val))
}

func Uint16p(key string, val *uint16) LogField {
	return LogField(zap.Uint16p(key, val))
}

func Uint32p(key string, val *uint32) LogField {
	return LogField(zap.Uint32p(key, val))
}

func Uint64p(key string, val *uint64) LogField {
	return LogField(zap.Uint64p(key, val))
}

func Intp(key string, val *int) LogField {
	return LogField(zap.Intp(key, val))
}

func Int8p(key string, val *int8) LogField {
	return LogField(zap.Int8p(key, val))
}

func Int16p(key string, val *int16) LogField {
	return LogField(zap.Int16p(key, val))
}

func Int32p(key string, val *int32) LogField {
	return LogField(zap.Int32p(key, val))
}

func Int64p(key string, val *int64) LogField {
	return LogField(zap.Int64p(key, val))
}

// 特殊类型转换
func NamedError(key string, err error) LogField {
	return LogField(zap.NamedError(key, err))
}

func Skip() LogField {
	return LogField(zap.Skip())
}

func Reflect(key string, val interface{}) LogField {
	return LogField(zap.Reflect(key, val))
}

func Namespace(key string) LogField {
	return LogField(zap.Namespace(key))
}

// 指针类型转换
func Stringp(key string, val *string) LogField {
	return LogField(zap.Stringp(key, val))
}

func Float32p(key string, val *float32) LogField {
	return LogField(zap.Float32p(key, val))
}

func Float64p(key string, val *float64) LogField {
	return LogField(zap.Float64p(key, val))
}

func Boolp(key string, val *bool) LogField {
	return LogField(zap.Boolp(key, val))
}

func Complex64p(key string, val *complex64) LogField {
	return LogField(zap.Complex64p(key, val))
}

func Complex128p(key string, val *complex128) LogField {
	return LogField(zap.Complex128p(key, val))
}

func Timep(key string, val *time.Time) LogField {
	return LogField(zap.Timep(key, val))
}

func Durationp(key string, val *time.Duration) LogField {
	return LogField(zap.Durationp(key, val))
}

// 其他类型转换
func Uintptr(key string, val uintptr) LogField {
	return LogField(zap.Uintptr(key, val))
}

func Uintptrp(key string, val *uintptr) LogField {
	return LogField(zap.Uintptrp(key, val))
}

func Stringer(key string, val fmt.Stringer) LogField {
	return LogField(zap.Stringer(key, val))
}

func Object(key string, val zapcore.ObjectMarshaler) LogField {
	return LogField(zap.Object(key, val))
}

func Inline(val zapcore.ObjectMarshaler) LogField {
	return LogField(zap.Inline(val))
}

func Dict(key string, val ...LogField) LogField {
	// 将 LogField 转换为 zap.Field
	zapFields := make([]zap.Field, len(val))
	for i, field := range val {
		zapFields[i] = zap.Field(field)
	}
	return LogField(zap.Dict(key, zapFields...))
}

func Stack(key string) LogField {
	return LogField(zap.Stack(key))
}

func StackSkip(key string, skip int) LogField {
	return LogField(zap.StackSkip(key, skip))
}

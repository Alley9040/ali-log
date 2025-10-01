package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	timeNow = time.Now().Format("2006010215")
)

func needRotation() bool {
	now := time.Now().Format("2006010215")
	if now != timeNow {
		timeNow = now
		return true
	}
	return false
}

func getFileName(level LogLevel) string {
	now := time.Now().Format("2006010215")
	return fmt.Sprintf("%s-%s.log", level.String(), now)
}

// SafeFileWriter 安全的文件写入器，支持原子性切换
type SafeFileWriter struct {
	file   *os.File
	mu     sync.RWMutex
	closed int32 // 使用原子操作标记是否已关闭
}

// Write 实现 io.Writer 接口
func (w *SafeFileWriter) Write(p []byte) (n int, err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// 检查文件是否已关闭
	if atomic.LoadInt32(&w.closed) == 1 || w.file == nil {
		return 0, fmt.Errorf("file already closed")
	}

	return w.file.Write(p)
}

// Sync 实现 zapcore.WriteSyncer 接口
func (w *SafeFileWriter) Sync() error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if atomic.LoadInt32(&w.closed) == 1 || w.file == nil {
		return fmt.Errorf("file already closed")
	}

	return w.file.Sync()
}

// Close 关闭文件写入器
func (w *SafeFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if atomic.LoadInt32(&w.closed) == 1 {
		return nil // 已经关闭
	}

	atomic.StoreInt32(&w.closed, 1)
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// SetFile 原子性地设置新的文件
func (w *SafeFileWriter) SetFile(file *os.File) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 关闭旧文件
	if w.file != nil {
		w.file.Close()
	}

	w.file = file
	atomic.StoreInt32(&w.closed, 0)
}

type log struct {
	cfg         *LogConfig
	logger      *zap.Logger
	fileWriters map[LogLevel]*SafeFileWriter
	mu          sync.RWMutex
	rotating    int32 // 标记是否正在滚动
}

func NewLogger(cfg *LogConfig) Log {
	impl := &log{
		cfg:         cfg,
		fileWriters: make(map[LogLevel]*SafeFileWriter),
	}

	// 初始化日志器
	impl.initLogger()

	return impl
}

// newBracketConsoleEncoder 创建控制台风格编码器，输出为：
// [yyyy-MM-dd HH:mm:ss:fff] [LEVEL] [caller] message messagedata
func newBracketConsoleEncoder() zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller: func(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + c.TrimmedPath() + "]")
		},
		EncodeLevel: func(lvl zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			name := lvl.CapitalString()
			if len(name) > 6 {
				name = name[:6]
			}
			if len(name) < 6 {
				name = strings.Repeat(" ", 6-len(name)) + name
			}
			enc.AppendString("[" + name + "]")
		},
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + t.Format("2006-01-02 15:04:05.000") + "]")
		},
		EncodeName:       zapcore.FullNameEncoder,
		ConsoleSeparator: " ",
	}
	return zapcore.NewConsoleEncoder(cfg)
}

// initLogger 初始化日志器
func (l *log) initLogger() {
	// 确保日志目录存在
	if err := os.MkdirAll(l.cfg.LogFileDir, 0755); err != nil {
		panic(fmt.Sprintf("创建日志目录失败: %v", err))
	}

	// 创建控制台与文件编码器（自定义行文本格式）
	consoleEncoder := newBracketConsoleEncoder()
	fileEncoder := newBracketConsoleEncoder()

	// 创建控制台输出
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), l.getZapLevelFromLogLevel(l.cfg.ConsoleLevel))
	// 控制台核心直接使用自定义时间与级别格式

	// 创建文件输出核心
	fileCore := l.createFileCore(fileEncoder)

	// 合并多个核心
	core := zapcore.NewTee(consoleCore, fileCore)

	// 创建logger，跳过一层包装方法（Debug/Info/Error等）所在的调用栈；
	// 仅在更高严重级别输出堆栈，避免 Error 级别打印堆栈；
	// Fatal 使用非退出钩子，避免 os.Exit(1)
	l.logger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithFatalHook(zapcore.WriteThenNoop),
	)
}

// createFileCore 创建文件输出核心
func (l *log) createFileCore(encoder zapcore.Encoder) zapcore.Core {
	// 为每个日志级别创建文件写入器
	cores := make([]zapcore.Core, 0, 6)

	levels := []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal, LogLevelPanic}

	for _, level := range levels {
		// 检查是否需要写入该级别的日志
		if level >= l.cfg.LogFileLevel {
			writer := l.getFileWriter(level)
			if writer != nil {
				// 仅写入“恰好等于该级别”的日志到对应文件；
				// panic 文件额外接收 DPanic 级别（避免进程终止时仍可记录到 panic 文件）
				targetLevel := l.getZapLevelFromLogLevel(level)
				levelOnly := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					if targetLevel == zapcore.PanicLevel {
						return lvl == zapcore.PanicLevel || lvl == zapcore.DPanicLevel
					}
					return lvl == targetLevel
				})
				core := zapcore.NewCore(encoder, writer, levelOnly)
				cores = append(cores, core)
			}
		}
	}

	// 如果没有文件核心，返回一个空的
	if len(cores) == 0 {
		return zapcore.NewNopCore()
	}

	// 合并所有文件核心
	return zapcore.NewTee(cores...)
}

// getFileWriter 获取文件写入器
func (l *log) getFileWriter(level LogLevel) *SafeFileWriter {
	l.mu.Lock()
	defer l.mu.Unlock()

	if writer, exists := l.fileWriters[level]; exists {
		return writer
	}

	// 创建新的文件写入器
	filePath := filepath.Join(l.cfg.LogFileDir, getFileName(level))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// 如果无法创建文件，返回nil，日志将只输出到控制台
		return nil
	}

	writer := &SafeFileWriter{file: file}
	l.fileWriters[level] = writer
	return writer
}

// getZapLevelFromLogLevel 将LogLevel转换为zap级别
func (l *log) getZapLevelFromLogLevel(level LogLevel) zapcore.Level {
	switch level {
	case LogLevelDebug:
		return zapcore.DebugLevel
	case LogLevelInfo:
		return zapcore.InfoLevel
	case LogLevelWarn:
		return zapcore.WarnLevel
	case LogLevelError:
		return zapcore.ErrorLevel
	case LogLevelFatal:
		return zapcore.FatalLevel
	case LogLevelPanic:
		return zapcore.PanicLevel
	default:
		return zapcore.DebugLevel
	}
}

// checkAndRotateLogs 检查并滚动日志
func (l *log) checkAndRotateLogs() {
	if !needRotation() {
		return
	}

	// 使用原子操作确保只有一个goroutine执行滚动
	if !atomic.CompareAndSwapInt32(&l.rotating, 0, 1) {
		return // 其他goroutine正在执行滚动，直接返回
	}
	defer atomic.StoreInt32(&l.rotating, 0)

	// 等待所有正在进行的日志写入完成
	l.mu.Lock()
	defer l.mu.Unlock()

	// 为每个级别创建新的文件并原子性地切换
	for level, writer := range l.fileWriters {
		if writer != nil {
			// 创建新的日志文件
			filePath := filepath.Join(l.cfg.LogFileDir, getFileName(level))
			newFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				// 如果无法创建新文件，保持使用旧文件
				continue
			}

			// 原子性地切换到新文件
			writer.SetFile(newFile)
		}
	}
}

// convertFields 转换LogField为zap.Field
func (l *log) convertFields(fields ...LogField) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Field(field)
	}
	return zapFields
}

// Debug 记录调试日志
func (l *log) Debug(msg string, fields ...LogField) {
	// 先检查是否需要滚动
	l.checkAndRotateLogs()

	// 如果正在滚动，等待完成
	for atomic.LoadInt32(&l.rotating) == 1 {
		time.Sleep(time.Millisecond)
	}

	l.logger.Debug(msg, l.convertFields(fields...)...)
}

// Info 记录信息日志
func (l *log) Info(msg string, fields ...LogField) {
	l.checkAndRotateLogs()
	for atomic.LoadInt32(&l.rotating) == 1 {
		time.Sleep(time.Millisecond)
	}
	l.logger.Info(msg, l.convertFields(fields...)...)
}

// Warn 记录警告日志
func (l *log) Warn(msg string, fields ...LogField) {
	l.checkAndRotateLogs()
	for atomic.LoadInt32(&l.rotating) == 1 {
		time.Sleep(time.Millisecond)
	}
	l.logger.Warn(msg, l.convertFields(fields...)...)
}

// Error 记录错误日志
func (l *log) Error(msg string, fields ...LogField) {
	l.checkAndRotateLogs()
	for atomic.LoadInt32(&l.rotating) == 1 {
		time.Sleep(time.Millisecond)
	}
	l.logger.Error(msg, l.convertFields(fields...)...)
}

// Fatal 记录致命错误日志
func (l *log) Fatal(msg string, fields ...LogField) {
	l.checkAndRotateLogs()
	for atomic.LoadInt32(&l.rotating) == 1 {
		time.Sleep(time.Millisecond)
	}
	l.logger.Fatal(msg, l.convertFields(fields...)...)
}

// Panic 记录恐慌日志
func (l *log) Panic(msg string, fields ...LogField) {
	l.checkAndRotateLogs()
	for atomic.LoadInt32(&l.rotating) == 1 {
		time.Sleep(time.Millisecond)
	}
	l.logger.Panic(msg, l.convertFields(fields...)...)
}

// Printf 格式化输出日志
func (l *log) Printf(format string, args ...interface{}) {
	l.checkAndRotateLogs()
	for atomic.LoadInt32(&l.rotating) == 1 {
		time.Sleep(time.Millisecond)
	}
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Close 关闭日志器并清理资源
func (l *log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var err error
	for level, writer := range l.fileWriters {
		if writer != nil {
			if closeErr := writer.Close(); closeErr != nil {
				err = closeErr
			}
			delete(l.fileWriters, level)
		}
	}

	// 清理旧日志文件
	l.cleanupOldLogs()

	return err
}

// cleanupOldLogs 清理超过最大保留时间的日志文件
func (l *log) cleanupOldLogs() {
	if l.cfg.LogFileMaxAge <= 0 {
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -l.cfg.LogFileMaxAge)

	// 遍历日志目录
	entries, err := os.ReadDir(l.cfg.LogFileDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 检查是否是日志文件
		if !isLogFile(entry.Name()) {
			continue
		}

		// 获取文件信息
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 如果文件超过最大保留时间，删除它
		if info.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(l.cfg.LogFileDir, entry.Name())
			os.Remove(filePath)
		}
	}
}

// isLogFile 检查文件名是否是日志文件
func isLogFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".log"
}

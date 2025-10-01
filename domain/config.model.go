package domain

// LogConfig 日志配置
type LogConfig struct {
	LogFileLevel   LogLevel `mapstructure:"logfile_level"`
	ConsoleLevel   LogLevel `mapstructure:"console_level"`
	LogFileDir     string   `mapstructure:"logfile_dir"`
	LogFileMaxSize int64    `mapstructure:"logfile_max_size"`
	LogFileMaxAge  int      `mapstructure:"logfile_max_age"`
}

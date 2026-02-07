package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// Logger represents a simple logger
type Logger struct {
	*log.Logger
	level   int
	module  string
	service string
}

// SugaredLogger wraps Logger with convenience methods
type SugaredLogger struct {
	*Logger
}

const (
	levelDebug = iota
	levelInfo
	levelWarn
	levelError
	levelFatal
)

var (
	defaultLogger *Logger
	mu            sync.Mutex
	logFile       *os.File
)

func init() {
	defaultLogger = &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  levelInfo,
	}
}

// GetLogConfig extracts log configuration from viper
func GetLogConfig(v *viper.Viper) *LogConfig {
	cfg := &LogConfig{
		Level:      "info",
		File:       "",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
	}
	if v != nil {
		_ = v.UnmarshalKey("log", cfg)
	}
	return cfg
}

// GetLogger returns a new logger instance
func GetLogger(service, module string, cfg *LogConfig) *Logger {
	mu.Lock()
	defer mu.Unlock()

	level := levelInfo
	if cfg != nil {
		switch cfg.Level {
		case "debug":
			level = levelDebug
		case "info":
			level = levelInfo
		case "warn":
			level = levelWarn
		case "error":
			level = levelError
		}
	}

	var writer io.Writer = os.Stdout

	if cfg != nil && cfg.File != "" {
		dir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(dir, 0755); err == nil {
			if f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				if logFile != nil {
					logFile.Close()
				}
				logFile = f
				writer = io.MultiWriter(os.Stdout, f)
			}
		}
	}

	return &Logger{
		Logger:  log.New(writer, "", log.LstdFlags),
		level:   level,
		module:  module,
		service: service,
	}
}

// Sugar returns a SugaredLogger
func (l *Logger) Sugar() *SugaredLogger {
	return &SugaredLogger{Logger: l}
}

func (l *Logger) log(level int, levelStr, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	l.Printf("[%s] [%s] [%s/%s] %s", timestamp, levelStr, l.service, l.module, msg)
}

// Debugf logs a debug message
func (l *SugaredLogger) Debugf(format string, args ...interface{}) {
	l.log(levelDebug, "DEBUG", format, args...)
}

// Infof logs an info message
func (l *SugaredLogger) Infof(format string, args ...interface{}) {
	l.log(levelInfo, "INFO", format, args...)
}

// Warnf logs a warning message
func (l *SugaredLogger) Warnf(format string, args ...interface{}) {
	l.log(levelWarn, "WARN", format, args...)
}

// Errorf logs an error message
func (l *SugaredLogger) Errorf(format string, args ...interface{}) {
	l.log(levelError, "ERROR", format, args...)
}

// Fatalf logs a fatal message and exits
func (l *SugaredLogger) Fatalf(format string, args ...interface{}) {
	l.log(levelFatal, "FATAL", format, args...)
	os.Exit(1)
}

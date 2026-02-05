package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel 日志级别
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// Logger 结构化日志器
type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// Config 日志配置
type Config struct {
	Level            LogLevel
	Environment      string // development, staging, production
	OutputPaths      []string
	ErrorOutputPaths []string
	EnableCaller     bool
	EnableStacktrace bool
	Encoding         string // json, console
	ServiceName      string
	ServiceVersion   string
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:            InfoLevel,
		Environment:      "development",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EnableCaller:     true,
		EnableStacktrace: true,
		Encoding:         "json",
		ServiceName:      "enterprise-platform",
		ServiceVersion:   "1.0.0",
	}
}

// NewLogger 创建日志器
func NewLogger(config *Config) (*Logger, error) {
	// 转换日志级别
	var level zapcore.Level
	switch config.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	var encoder zapcore.Encoder
	if config.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建输出
	var outputPaths []string
	if len(config.OutputPaths) > 0 {
		outputPaths = config.OutputPaths
	} else {
		outputPaths = []string{"stdout"}
	}

	// 构建zap配置
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       config.Environment == "development",
		DisableCaller:     !config.EnableCaller,
		DisableStacktrace: !config.EnableStacktrace,
		Encoding:          config.Encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       outputPaths,
		ErrorOutputPaths:  config.ErrorOutputPaths,
		InitialFields: map[string]interface{}{
			"service": config.ServiceName,
			"version": config.ServiceVersion,
			"env":     config.Environment,
		},
	}

	// 构建logger
	zapLogger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// Debug 调试日志
func (l *Logger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, toZapFields(fields)...)
}

// Info 信息日志
func (l *Logger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, toZapFields(fields)...)
}

// Warn 警告日志
func (l *Logger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, toZapFields(fields)...)
}

// Error 错误日志
func (l *Logger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, toZapFields(fields)...)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, toZapFields(fields)...)
}

// Debugf 格式化调试日志
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Infof 格式化信息日志
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warnf 格式化警告日志
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Errorf 格式化错误日志
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatalf 格式化致命错误日志
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// WithContext 带上下文的日志器
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := extractContextFields(ctx)
	return &Logger{
		logger: l.logger.With(toZapFields(fields)...),
		sugar:  l.sugar.With(toZapFields(fields)...),
	}
}

// WithFields 添加字段
func (l *Logger) WithFields(fields ...Field) *Logger {
	return &Logger{
		logger: l.logger.With(toZapFields(fields)...),
		sugar:  l.sugar.With(toZapFields(fields)...),
	}
}

// Sync 刷新日志
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// String 字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 64位整数字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 浮点数字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration 时长字段
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Error 错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Any 任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// toZapFields 转换为zap字段
func toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(f.Key, v))
		case int:
			zapFields = append(zapFields, zap.Int(f.Key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(f.Key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(f.Key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(f.Key, v))
		case time.Duration:
			zapFields = append(zapFields, zap.Duration(f.Key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		default:
			zapFields = append(zapFields, zap.Any(f.Key, v))
		}
	}
	return zapFields
}

// extractContextFields 从上下文提取字段
func extractContextFields(ctx context.Context) []Field {
	fields := make([]Field, 0)

	// 提取常见上下文字段
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, String("request_id", fmt.Sprint(requestID)))
	}

	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		fields = append(fields, String("tenant_id", fmt.Sprint(tenantID)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, String("user_id", fmt.Sprint(userID)))
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, String("trace_id", fmt.Sprint(traceID)))
	}

	if spanID := ctx.Value("span_id"); spanID != nil {
		fields = append(fields, String("span_id", fmt.Sprint(spanID)))
	}

	return fields
}

// Global logger
var globalLogger *Logger

func init() {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	globalLogger = logger
}

// SetGlobalLogger 设置全局日志器
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger 获取全局日志器
func GetGlobalLogger() *Logger {
	return globalLogger
}

// 全局快捷方法
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	globalLogger.Fatal(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	globalLogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	globalLogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	globalLogger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	globalLogger.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	globalLogger.Fatalf(template, args...)
}

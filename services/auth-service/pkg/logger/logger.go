package logger

import (
	"auth-service/internal/config"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(cfg config.LoggerConfig) (*Logger, error) {
	var zapCfg zap.Config

	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	if err := zapCfg.Level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	zapCfg.Encoding = cfg.Format

	zapCfg.OutputPaths = cfg.OutputPaths
	zapCfg.ErrorOutputPaths = cfg.ErrorOutputPaths

	switch strings.ToLower(cfg.TimeFormat) {
	case "iso8601":
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "rfc3339":
		zapCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	case "epoch":
		zapCfg.EncoderConfig.EncodeTime = zapcore.EpochTimeEncoder
	case "millis":
		zapCfg.EncoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
	default:
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if cfg.EnableStacktrace {
		zapCfg.EncoderConfig.StacktraceKey = "stacktrace"
	} else {
		zapCfg.EncoderConfig.StacktraceKey = ""
	}

	coreLogger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{coreLogger.Sugar()}, nil
}

// Sync аккуратно закрывает логгер
func (l *Logger) Sync() error {
	return l.SugaredLogger.Desugar().Sync()
}

// With создает новый логгер с дополнительными полями
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{l.SugaredLogger.With(args...)}
}

// WithFields создает логгер с структурированными полями
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return l.With(args...)
}

// WithError создает логгер с добавленной ошибкой
func (l *Logger) WithError(err error) *Logger {
	return l.With("error", err.Error())
}

// Debug логирование на уровне Debug
func (l *Logger) Debug(args ...interface{}) {
	l.SugaredLogger.Debug(args...)
}

// Debugf логирование с форматированием на уровне Debug
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.SugaredLogger.Debugf(template, args...)
}

// Info логирование на уровне Info
func (l *Logger) Info(args ...interface{}) {
	l.SugaredLogger.Info(args...)
}

// Infof логирование с форматированием на уровне Info
func (l *Logger) Infof(template string, args ...interface{}) {
	l.SugaredLogger.Infof(template, args...)
}

// Warn логирование на уровне Warn
func (l *Logger) Warn(args ...interface{}) {
	l.SugaredLogger.Warn(args...)
}

// Warnf логирование с форматированием на уровне Warn
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.SugaredLogger.Warnf(template, args...)
}

// Error логирование на уровне Error
func (l *Logger) Error(args ...interface{}) {
	l.SugaredLogger.Error(args...)
}

// Errorf логирование с форматированием на уровне Error
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.SugaredLogger.Errorf(template, args...)
}

// Fatal логирование на уровне Fatal с завершением программы
func (l *Logger) Fatal(args ...interface{}) {
	l.SugaredLogger.Fatal(args...)
	os.Exit(1)
}

// Fatalf логирование с форматированием на уровне Fatal с завершением программы
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.SugaredLogger.Fatalf(template, args...)
	os.Exit(1)
}

// Panic логирование на уровне Panic
func (l *Logger) Panic(args ...interface{}) {
	l.SugaredLogger.Panic(args...)
}

// Panicf логирование с форматированием на уровне Panic
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.SugaredLogger.Panicf(template, args...)
}

// LogDuration измеряет время выполнения функции
func (l *Logger) LogDuration(message string, start time.Time) {
	duration := time.Since(start)
	l.Infof("%s took %s", message, duration)
}

// LogMethodCall логирует вызов метода с автоматическим определением caller
func (l *Logger) LogMethodCall() {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		l.Error("Could not get caller info")
		return
	}

	function := runtime.FuncForPC(pc).Name()
	l.Debugf("Method called: %s at %s:%d", function, file, line)
}

// IsDebugEnabled проверяет, включено ли debug-логирование
func (l *Logger) IsDebugEnabled() bool {
	// Для простоты проверяем уровень логирования
	// В реальной реализации нужно парсить уровень из конфига
	return l.SugaredLogger.Desugar().Core().Enabled(zapcore.DebugLevel)
}

// GetZapLogger возвращает оригинальный zap логгер
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.SugaredLogger.Desugar()
}

// Named создает именованный логгер
func (l *Logger) Named(name string) *Logger {
	return &Logger{l.SugaredLogger.Named(name)}
}

// WithOptions создает логгер с дополнительными опциями
func (l *Logger) WithOptions(opts ...zap.Option) *Logger {
	return &Logger{l.SugaredLogger.Desugar().WithOptions(opts...).Sugar()}
}

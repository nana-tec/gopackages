package ntlogger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var zapSinLogger *zap.SugaredLogger
var once sync.Once

type zapLogger struct {
	cfg    LogConfig
	logger *zap.SugaredLogger
}

var zapLogLevelMapping = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

func newZapLogger(cfg LogConfig) *zapLogger {
	logger := &zapLogger{cfg: cfg}
	logger.Init()
	return logger
}

func (l *zapLogger) getLogLevel() zapcore.Level {
	level, exists := zapLogLevelMapping[l.cfg.Level]
	if !exists {
		return zapcore.DebugLevel
	}
	return level
}

// newResource creates a new OTEL resource with the service name and version.

func (l *zapLogger) Init() {
	once.Do(func() {
		fileName := fmt.Sprintf("%s%s.%s", l.cfg.FilePath, time.Now().Format("2006-01-02"), "log")
		//fileName := fmt.Sprintf("%s%s.%s", "./logs/", time.Now().Format("2006-01-02"), "log")
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   fileName,
			MaxSize:    1,
			MaxAge:     20,
			LocalTime:  true,
			MaxBackups: 5,
			Compress:   true,
		})

		config := zap.NewProductionEncoderConfig()
		config.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(config),
			w,
			l.getLogLevel(),
		)

		logger := zap.New(core, zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zapcore.ErrorLevel),
		).Sugar()

		zapSinLogger = logger.With("AppName", l.cfg.AppName, "AppServiceName", l.cfg.AppServiceName, "AppNameSpace", l.cfg.AppNameSpace, "Environment", l.cfg.Environment, "pid", os.Getpid())
	})

	l.logger = zapSinLogger
}

// ctx context.Context, code string, msg string, extra map[ExtraKey]interface{}
func (l *zapLogger) Debug(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogInfo(code, extra)

	l.logger.Debugw(msg, params...)
}

func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.logger.Debugf(template, args)
}

func (l *zapLogger) Info(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogInfo(code, extra)
	l.logger.Infow(msg, params...)
}

func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.logger.Infof(template, args)
}

func (l *zapLogger) Warn(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogInfo(code, extra)
	l.logger.Warnw(msg, params...)
}

func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.logger.Warnf(template, args)
}

func (l *zapLogger) Error(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogInfo(code, extra)
	l.logger.Errorw(msg, params...)
}

func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.logger.Errorf(template, args)
}

func (l *zapLogger) Fatal(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogInfo(code, extra)
	l.logger.Fatalw(msg, params...)
}

func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.logger.Fatalf(template, args)
}

func prepareLogInfo(code string, extra map[ExtraKey]interface{}) []interface{} {
	if extra == nil {
		extra = make(map[ExtraKey]interface{})
	}
	extra["code"] = code

	return logParamsToZapParams(extra)
}

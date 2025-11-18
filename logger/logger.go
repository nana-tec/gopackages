package ntlogger

import (
	"context"
)

type Logger interface {
	Init()

	Debug(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{})
	Debugf(template string, args ...interface{})

	Info(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{})
	Infof(template string, args ...interface{})

	Warn(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{})
	Warnf(template string, args ...interface{})

	Error(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{})
	Errorf(template string, args ...interface{})

	Fatal(ctx context.Context, code string, msg string, extra map[ExtraKey]interface{})
	Fatalf(template string, args ...interface{})
}

func NewLogger(cfg LogConfig) Logger {
	/*if cfg.Logger.Logger == "zap" {
		return newZapLogger(cfg)
	} else if cfg.Logger.Logger == "zerolog" {
		return newZeroLogger(cfg)
	}
	panic("logger not supported")
	*/

	return newZapLogger(cfg)
}

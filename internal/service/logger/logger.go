package logger

import (
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const traceIdField = "traceID"

type Logger interface {
	Debug(msg string, data ...LogDataItem)
	Info(msg string, data ...LogDataItem)
	Warn(msg string, data ...LogDataItem)
	Error(msg string, data ...LogDataItem)
	Fatal(msg string, data ...LogDataItem)
	SetTraceId(id string)
}

type LogDataItem struct {
	Key   string
	Value any
}

type zapLogger struct {
	zap     *zap.Logger
	traceId string
}

func (z *zapLogger) dataToFields(data ...LogDataItem) []zap.Field {
	fields := make([]zap.Field, 0, len(data))

	for _, item := range data {
		fields = append(fields, zap.Any(item.Key, item.Value))
	}

	if z.traceId != "" {
		fields = append(fields, zap.String(traceIdField, z.traceId))
	}

	return fields
}

func (z *zapLogger) Debug(msg string, data ...LogDataItem) {
	z.zap.Debug(msg, z.dataToFields(data...)...)
}

func (z *zapLogger) Info(msg string, data ...LogDataItem) {
	z.zap.Info(msg, z.dataToFields(data...)...)
}

func (z *zapLogger) Warn(msg string, data ...LogDataItem) {
	z.zap.Warn(msg, z.dataToFields(data...)...)
}

func (z *zapLogger) Error(msg string, data ...LogDataItem) {
	z.zap.Error(msg, z.dataToFields(data...)...)
}

func (z *zapLogger) Fatal(msg string, data ...LogDataItem) {
	z.zap.Fatal(msg, z.dataToFields(data...)...)
}

func (z *zapLogger) SetTraceId(traceId string) {
	z.traceId = traceId
}

func NewLogger(level string, env string) (Logger, error) {
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewDevelopmentConfig()

	if env == config.EnvProd {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)
	cfg.DisableCaller = true

	zap, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		zap: zap,
	}, nil
}

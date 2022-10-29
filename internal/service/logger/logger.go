package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const traceIdField = "traceID"

var (
	logger   *zap.Logger
	zapLevel zap.AtomicLevel
	traceID  string
)

func init() {
	zapLevel = zap.NewAtomicLevel()

	logger = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zapLevel,
	))
}

func SetLevel(level string) {
	newLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		log.Printf("ошибка установки уровня логов: %s", err.Error())
	}

	zapLevel.SetLevel(newLevel)
}

type LogDataItem struct {
	Key   string
	Value any
}

func dataToFields(data ...LogDataItem) []zap.Field {
	fields := make([]zap.Field, 0, len(data))

	for _, item := range data {
		fields = append(fields, zap.Any(item.Key, item.Value))
	}

	if traceID != "" {
		fields = append(fields, zap.String(traceIdField, traceID))
	}

	return fields
}

func Debug(msg string, data ...LogDataItem) {
	logger.Debug(msg, dataToFields(data...)...)
}

func Info(msg string, data ...LogDataItem) {
	logger.Info(msg, dataToFields(data...)...)
}

func Warn(msg string, data ...LogDataItem) {
	logger.Warn(msg, dataToFields(data...)...)
}

func Error(msg string, data ...LogDataItem) {
	logger.Error(msg, dataToFields(data...)...)
}

func Fatal(msg string, data ...LogDataItem) {
	logger.Fatal(msg, dataToFields(data...)...)
}

func SetTraceId(id string) {
	traceID = id
}

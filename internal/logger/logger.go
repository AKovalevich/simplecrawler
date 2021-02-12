package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	err := config.Level.UnmarshalText([]byte(level))
	if err != nil || len(level) == 0 {
		return nil, fmt.Errorf("invalid debug level %s\n", level)
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

package utils

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewCustomLogger(level zapcore.Level, outputToFiles bool) (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.CallerKey = ""

	outputPaths := []string{"stdout"}
	errorOutputPaths := []string{"stderr"}

	if outputToFiles {
		outputPaths = append(outputPaths, "./logs.log")
		errorOutputPaths = append(errorOutputPaths, "./errors.log")
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:          "console",
		EncoderConfig:     encoderConfig,
		OutputPaths:       outputPaths,
		ErrorOutputPaths:  errorOutputPaths,
		DisableStacktrace: true,
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create logger %w", err)
	}

	return logger, nil
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(debug bool) *zap.Logger {
	startTime := time.Now().Format("2006-01-02_15-04-05")
	logDir := filepath.Join("logs", startTime)

	if err := os.MkdirAll(logDir, 0750); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}

	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	allFile, _ := os.Create(filepath.Join(logDir, "all.json"))   //nolint:gosec // log path constructed internally
	errFile, _ := os.Create(filepath.Join(logDir, "error.json")) //nolint:gosec // log path constructed internally
	httpFile, _ := os.Create(filepath.Join(logDir, "http.json")) //nolint:gosec // log path constructed internally

	allLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l >= zapcore.InfoLevel })
	errLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l >= zapcore.ErrorLevel })

	var cores []zapcore.Core

	cores = append(
		cores,
		zapcore.NewCore(fileEncoder, zapcore.AddSync(allFile), allLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(errFile), errLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(httpFile), allLevel),
	)

	if debug {
		cores = append(
			cores,
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), allLevel),
		)
	}

	core := zapcore.NewTee(cores...)

	return zap.New(core, zap.AddCaller())
}

package main

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger = slog.New(handler)
}

func LogInfo(enMsg, jpMsg string, args ...any) {
	msg := enMsg + " / " + jpMsg
	logger.Info(msg, args...)
}

func LogWarn(enMsg, jpMsg string, args ...any) {
	msg := enMsg + " / " + jpMsg
	logger.Warn(msg, args...)
}

func LogError(enMsg, jpMsg string, args ...any) {
	msg := enMsg + " / " + jpMsg
	logger.Error(msg, args...)
}

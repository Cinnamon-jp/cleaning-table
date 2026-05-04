package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestCustomLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	customLogger := NewCustomLogger(logger)

	title := "TestTitle"
	enBody := "This is a test message."
	jaBody := "これはテストメッセージです。"

	customLogger.Info(title, enBody, jaBody)

	output := buf.String()
	if !strings.Contains(output, "level=INFO") {
		t.Errorf("Expected Info level, got %s", output)
	}
	if !strings.Contains(output, "msg=TestTitle") {
		t.Errorf("Expected msg=TestTitle, got %s", output)
	}
	if !strings.Contains(output, "en=\"This is a test message.\"") {
		t.Errorf("Expected en text, got %s", output)
	}
	if !strings.Contains(output, "ja=これはテストメッセージです。") {
		t.Errorf("Expected ja text, got %s", output)
	}
}

func TestCustomLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	customLogger := NewCustomLogger(logger)

	title := "ErrorTitle"
	enBody := "This is an error message."
	jaBody := "これはエラーメッセージです。"

	customLogger.Error(title, enBody, jaBody)

	output := buf.String()
	if !strings.Contains(output, "level=ERROR") {
		t.Errorf("Expected Error level, got %s", output)
	}
	if !strings.Contains(output, "msg=ErrorTitle") {
		t.Errorf("Expected msg=ErrorTitle, got %s", output)
	}
	if !strings.Contains(output, "en=\"This is an error message.\"") {
		t.Errorf("Expected en text, got %s", output)
	}
	if !strings.Contains(output, "ja=これはエラーメッセージです。") {
		t.Errorf("Expected ja text, got %s", output)
	}
}

func TestCustomLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	customLogger := NewCustomLogger(logger)

	title := "WarnTitle"
	enBody := "This is a warning message."
	jaBody := "これは警告メッセージです。"

	customLogger.Warn(title, enBody, jaBody)

	output := buf.String()
	if !strings.Contains(output, "level=WARN") {
		t.Errorf("Expected Warn level, got %s", output)
	}
	if !strings.Contains(output, "msg=WarnTitle") {
		t.Errorf("Expected msg=WarnTitle, got %s", output)
	}
	if !strings.Contains(output, "en=\"This is a warning message.\"") {
		t.Errorf("Expected en text, got %s", output)
	}
	if !strings.Contains(output, "ja=これは警告メッセージです。") {
		t.Errorf("Expected ja text, got %s", output)
	}
}

func TestCustomLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	// Debugレベルを出力するようにオプションを設定
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger := slog.New(handler)
	customLogger := NewCustomLogger(logger)

	title := "DebugTitle"
	enBody := "This is a debug message."
	jaBody := "これはデバッグメッセージです。"

	customLogger.Debug(title, enBody, jaBody)

	output := buf.String()
	if !strings.Contains(output, "level=DEBUG") {
		t.Errorf("Expected Debug level, got %s", output)
	}
	if !strings.Contains(output, "msg=DebugTitle") {
		t.Errorf("Expected msg=DebugTitle, got %s", output)
	}
	if !strings.Contains(output, "en=\"This is a debug message.\"") {
		t.Errorf("Expected en text, got %s", output)
	}
	if !strings.Contains(output, "ja=これはデバッグメッセージです。") {
		t.Errorf("Expected ja text, got %s", output)
	}
}

func TestNewCustomLogger_Default(t *testing.T) {
	customLogger := NewCustomLogger(nil)
	if customLogger.logger == nil {
		t.Errorf("Expected default logger to be set, but got nil")
	}
}

func TestLogger_Initialized(t *testing.T) {
	if logger == nil {
		t.Fatal("logger should be initialized by init function")
	}
	if logger.logger == nil {
		t.Error("logger.logger should not be nil")
	}
}

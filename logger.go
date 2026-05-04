package main

import (
	"log/slog"
)

// Logger はアプリケーション全体で共有されるロガーのインスタンスです。
// logger.go の init 関数によって自動的に初期化されます。
var logger *CustomLogger

func init() {
	// デフォルトの slog.Logger を使用して CustomLogger を初期化します。
	logger = NewCustomLogger(nil)
}

// CustomLogger は独自のフォーマットでログ出力を提供するロガーです。
type CustomLogger struct {
	logger *slog.Logger
}

// NewCustomLogger は新しい CustomLogger のインスタンスを生成して返します。
// l に nil が渡された場合はデフォルトのロガーを使用します。
func NewCustomLogger(l *slog.Logger) *CustomLogger {
	if l == nil {
		l = slog.Default()
	}
	return &CustomLogger{logger: l}
}

// Info は Info レベルでログを出力します。
func (c *CustomLogger) Info(title string, enBody string, jaBody string) {
	c.logger.Info(title, slog.String("en", enBody), slog.String("ja", jaBody))
}

// Error は Error レベルでログを出力します。
func (c *CustomLogger) Error(title string, enBody string, jaBody string) {
	c.logger.Error(title, slog.String("en", enBody), slog.String("ja", jaBody))
}

// Warn は Warn レベルでログを出力します。
func (c *CustomLogger) Warn(title string, enBody string, jaBody string) {
	c.logger.Warn(title, slog.String("en", enBody), slog.String("ja", jaBody))
}

// Debug は Debug レベルでログを出力します。
func (c *CustomLogger) Debug(title string, enBody string, jaBody string) {
	c.logger.Debug(title, slog.String("en", enBody), slog.String("ja", jaBody))
}
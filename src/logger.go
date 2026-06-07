package src

import (
	"log/slog"
)

type LogLabel int

const (
	Unknown LogLabel = iota
	Info
	Warn
	Error
)

// Logger は英語と日本語を併記してログを出力する (文字化け対処用)
func Logger(
	logType LogLabel,
	Where string,
	ENBody string,
	JPBody string,
) {
	switch logType {
	case Info:
		slog.Info(Where + ": " + ENBody + " / " + JPBody)
	case Warn:
		slog.Warn(Where + ": " + ENBody + " / " + JPBody)
	case Error:
		slog.Error(Where + ": " + ENBody + " / " + JPBody)
	default:
	}
}

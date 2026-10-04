// Package main は掃除当番表生成ツールのCLIエントリポイントです。
package main

import (
	"log/slog"
	"os"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	cli := &CLI{
		InStream:  os.Stdin,
		OutStream: os.Stdout,
		ErrStream: os.Stderr,
	}

	return cli.Run(os.Args[1:])
}

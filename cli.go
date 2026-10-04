package main

import (
	"errors"
	"flag"
	"io"
)

type CLI struct {
	InStream  io.Reader
	OutStream io.Writer
	ErrStream io.Writer
}

func (c *CLI) Run(args []string) error {
	flags := flag.NewFlagSet("cleaning-table", flag.ContinueOnError)
	flags.SetOutput(c.ErrStream)

	configPath := flags.String("config", "config.yaml", "path to config file (yaml)")
	outputPath := flags.String("output", "output.pdf", "path to output file (pdf)")
	reversePath := flags.String("reverse", "reverse.txt", "path to reverse output file (txt)")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	_ = configPath
	_ = outputPath
	_ = reversePath

	return nil
}

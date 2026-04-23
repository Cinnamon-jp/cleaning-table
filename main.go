package main

import (
	"log/slog"

	"github.com/xuri/excelize/v2"
)

func main() {
	err := run()
	if err != nil {
		slog.Error("Error: "+err.Error())
	}
}

func run() error {
	slog.Info("Hello, World!")
	f, err := excelize.OpenFile("test.xlsx")
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("When closing .xlsx file")
		}
	}()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}
	for i, row := range rows {
		for j, cell := range row {
			slog.Info("Cell", "row_number", i, "col_number", j, "value", cell)
		}
	}

	return nil
}

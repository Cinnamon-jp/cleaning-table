package main

import (
	"os"
	"path/filepath"
)

func main() {
	
}

func run() error {
	
}

func findExcelFiles() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".xlsx" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
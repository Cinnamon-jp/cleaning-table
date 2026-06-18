package excel

import (
	"fmt"
	"os"
	"strings"
)

// GetExcel は指定されたパス内にある .xlsx ファイルのファイル名のスライスを返します。
func GetExcel(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading directory / ディレクトリの読み込み中: %w", err)
	}

	var xlsxFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".xlsx") {
			xlsxFiles = append(xlsxFiles, entry.Name())
		}
	}

	return xlsxFiles, nil
}

package src

import (
	"os"
	"strings"
)

// GetExcel は指定されたパス内にある .xlsx ファイルのファイル名のスライスを返します。
func GetExcel(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		Logger(Error, "src/getExcel.go/GetExcel()/os.ReadDir()", "Failed to read directory", "ディレクトリの読み込みに失敗しました")
		return nil, err
	}

	var xlsxFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".xlsx") {
			xlsxFiles = append(xlsxFiles, entry.Name())
		}
	}

	return xlsxFiles, nil
}

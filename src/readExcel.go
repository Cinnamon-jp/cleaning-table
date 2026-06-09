// Package src はプログラムを構成する部品に分割して格納する
package src

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ReadExcel は指定されたパスのExcelファイルを開き、選択したシートの全行のデータを二次元配列として返す
func ReadExcel(path, sheetName string) ([][]string, error) {
	// Excelファイルを開く
	f, err := excelize.OpenFile(path)
	if err != nil {
		Logger(Error, "src/readExcel.go/ReadExcel()/excelize.OpenFile()", "Failed to open Excel file", "Excelファイルのオープンに失敗しました")
		return nil, fmt.Errorf("excelファイルのオープンに失敗しました (path: %s): %w", path, err)
	}
	// 関数終了時にファイルを確実に閉じる
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			Logger(Error, "src/readExcel.go/ReadExcel()/f.Close()", "Failed to close Excel file", "ファイルのクローズに失敗しました")
		}
	}()

	// 指定したシートの全行を取得する
	rows, err := f.GetRows(sheetName)
	if err != nil {
		Logger(Error, "src/readExcel.go/ReadExcel()/f.GetRows()", "Failed to get rows from sheet", "シートからの行データ取得に失敗しました")
		return nil, fmt.Errorf("シートからの行データ取得に失敗しました(sheetName: %s): %w", sheetName, err)
	}

	return rows, nil
}

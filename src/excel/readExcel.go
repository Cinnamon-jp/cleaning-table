package excel

import (
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ReadExcel は指定されたパスのExcelファイルを開き、選択したシートの全行のデータを二次元配列として返す
func ReadExcel(path, sheetName string) (rowsData [][]string, err error) {
	// Excelファイルを開く
	f, openErr := excelize.OpenFile(path)
	if openErr != nil {
		return nil, fmt.Errorf("opening the Excel file / Excelファイルの読み込み中: %w", openErr)
	}
	// 関数終了時にファイルを確実に閉じる
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("closing the Excel file / Excelファイルのクローズ中: %w", closeErr)
			} else {
				err = errors.Join(err, fmt.Errorf("closing the Excel file / Excelファイルのクローズ中: %w", closeErr))
			}
			rowsData = nil
		}
	}()

	// 指定したシートの全行を取得する
	rows, getErr := f.GetRows(sheetName)
	if getErr != nil {
		return nil, fmt.Errorf("getting rows from sheet / シートからの行データ取得中: %w", getErr)
	}

	return rows, nil
}

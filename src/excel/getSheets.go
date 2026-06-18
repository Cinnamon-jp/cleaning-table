package excel

import (
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// GetSheets は指定されたExcelファイルからシートの一覧リストを取得します。
func GetSheets(filePath string) (sheets []string, err error) {
	f, openErr := excelize.OpenFile(filePath)
	if openErr != nil {
		return nil, fmt.Errorf("opening the Excel file / Excelファイルの読み込み中: %w", openErr)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("closing the Excel file / Excelファイル閉じ中: %w", closeErr)
			} else {
				err = errors.Join(err, fmt.Errorf("closing the Excel file / Excelファイル閉じ中: %w", closeErr))
			}
			sheets = nil
		}
	}()

	return f.GetSheetList(), nil
}

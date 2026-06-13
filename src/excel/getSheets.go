package excel

import (
	"cleaning-table/src/util"

	"github.com/xuri/excelize/v2"
)

// GetSheets は指定されたExcelファイルからシートの一覧リストを取得します。
func GetSheets(filePath string) ([]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		util.Logger(util.Error, "getSheets.go/GetSheets()/excelize.OpenFile()", "Error when executing OpenFile()", "Excelファイルの読み込み中にエラーが発生しました")
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			util.Logger(util.Error, "getSheets.go/GetSheets()/f.Close()", "Error when executing Close()", "Excelファイルのクローズ中にエラーが発生しました")
		}
	}()

	return f.GetSheetList(), nil
}

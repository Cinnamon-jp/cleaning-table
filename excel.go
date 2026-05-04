package main

import (
	"github.com/xuri/excelize/v2"
)

// GetExcelData は指定されたパスのエクセルファイルを読み込み、内容を2次元スライスとして返します。
func GetExcelData(path string) ([][]string, error) {
	return readExcelFile(path)
}

// readExcelFile は指定されたパスのエクセルファイルを読み込み、内容を2次元スライスとして返します。
func readExcelFile(path string) ([][]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		logger.Error(
			"excel.go: excelize.OpenFile()",
			"Couldn't open excel file",
			"エクセルファイルを開くことができませんでした",
		)
		return nil, err
	}
	defer func() {
			if err := f.Close(); err != nil {
				logger.Error(
					"excel.go: excelize.Close()",
					"Couldn't close excel file",
					"エクセルファイルを閉じることができませんでした",
				)
			}
	}()

	// 使用するシートを選択
	sheetList := f.GetSheetList()
	var sheetToUse string
	if len(sheetList) != 1 {
		sheetToUse, err = ChooseOne(
			"Choose the sheet you want to use. / 使用するシートを選択してください。",
			sheetList,
		)
		if err != nil {
			logger.Error(
				"excel.go: ChooseOne()",
				"Couldn't choose sheet",
				"シートを選択できませんでした",
			)
			return nil, err
		}
	} else {
		sheetToUse = sheetList[0]
	}

	rows, err := f.GetRows(sheetToUse)
	if err != nil {
		logger.Error(
			"excel.go: excelize.GetRows()",
			"Couldn't get rows",
			"行を取得できませんでした",
		)
		return nil, err
	}

	return rows, nil
}

// checkExcelSyntax はエクセルファイルの文法をチェックします。
// func checkExcelSyntax(data [][]*string) error {
	
// }
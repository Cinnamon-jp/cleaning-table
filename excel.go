package main

import (
	"cleaning-table/util"
	"errors"

	"github.com/xuri/excelize/v2"
)

type UnfoldedExcelData struct {
	roomNumbers [][]int
	tasks [][]string
}

// getExcelData は指定されたパスのエクセルファイルを読み込み、内容を2次元スライスとして返します。
func getExcelData(path string) ([][]string, error) {
	return readExcelFile(path)
}

// readExcelFile は指定されたパスのエクセルファイルを読み込み、内容を2次元スライスとして返します。
func readExcelFile(path string) ([][]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		util.Logger.Error(
			"excel.go: excelize.OpenFile()",
			"Couldn't open excel file",
			"エクセルファイルを開くことができませんでした",
		)
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
				util.Logger.Error(
					"excel.go: excelize.Close()",
					"Couldn't close excel file",
					"エクセルファイルを閉じることができませんでした",
				)
			}
	}()

	// シートリストを取得
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		util.Logger.Error(
			"excel.go: excelize.GetSheetList()",
			"No sheets found in excel file",
			"エクセルファイルにシートが見つかりませんでした",
		)
		return nil, errors.New("no sheets found in excel file")
	}

	// シートリストの数が1以上なら使用するシートを選択
	sheetToUse := ""
	if len(sheetList) == 1 {
		sheetToUse = sheetList[0]
	} else {
		sheetToUse, err = util.ChooseOne(
			"Choose the sheet you want to use. / 使用するシートを選択してください。",
			sheetList,
		)
		if err != nil {
			util.Logger.Error(
				"excel.go: ChooseOne()",
				"Couldn't choose sheet",
				"シートを選択できませんでした",
			)
			return nil, err
		}
	}

	rows, err := f.GetRows(sheetToUse)
	if err != nil {
		util.Logger.Error(
			"excel.go: excelize.GetRows()",
			"Couldn't get rows",
			"行を取得できませんでした",
		)
		return nil, err
	}

	return rows, nil
}

// unfoldExcelData はででｄでデータの部屋番号範囲、役職数を展開する
func unfoldExcelData(data [][]string) (UnfoldedExcelData, error) {
	
}


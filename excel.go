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
func getExcelData(path string) (*UnfoldedExcelData, error) {
	excel, err := readExcelFile(path)
	if err != nil {
		util.Logger.Error(
			"excel.go: readExcelFile()",
			"Couldn't read excel file",
			"エクセルファイルの内容を読み込むことができませんでした",
		)
		return nil, err
	}
	
	// コメント行を削除
	excel = deleteCommentRow(excel)
	
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

// deleteCommentRow はエクセルデータのコメント行を削除する
func deleteCommentRow(data [][]string) [][]string {
	var result [][]string
	for _, row := range data {
		if len(row) > 0 && row[0] == "#" {
			continue
		}
		result = append(result, row)
	}
	return result
}

// unfoldExcelData は部屋番号範囲と役職数を展開する
func unfoldExcelData(data [][]string) (UnfoldedExcelData, error) {
	if len(data) == 0 {
		util.Logger.Error(
			"excel.go: unfoldExcelData()",
			"Empty data",
			"データが空です",
		)
		return UnfoldedExcelData{}, errors.New("empty data")
	}

	roomNumberStrings := data[0]
	taskStrings := data[1:]
}


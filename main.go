package main

import (
	"cleaning-table/src"
	"fmt"
)

func main() {
	if err := run(); err != nil {
		src.Logger(src.Error, "main.go/main()/run()", "Error when executing run()", "run()の実行中にエラーが発生しました")
	}
}

func run() error {
	// Excelファイルリストの取得
	var excelFiles []string
	var err error
	if excelFiles, err = src.GetExcel("."); err != nil {
		src.Logger(src.Error, "main.go/run()/src.GetExcel()", "Error when executing GetExcel()", "GetExcel()の実行中にエラーが発生しました")
		return err
	}

	fmt.Printf("カレントディレクトリにあるエクセルファイル一覧: %v\n", excelFiles) // dev

	// 使用するExcelファイルを選択する
	var selectedExcelFile string
	if len(excelFiles) > 1 {
		if selectedExcelFile, err = src.Select("使用するExcelファイルを選択してください", excelFiles); err != nil {
			src.Logger(src.Error, "main.go/run()/src.Select()", "Error when executing Select()", "Select()の実行中にエラーが発生しました")
			return err
		}
	} else {
		selectedExcelFile = excelFiles[0]
	}

	fmt.Printf("選択されたExcelファイル%v\n", selectedExcelFile) // dev

	// シートの一覧リストを取得
	sheetList, err := src.GetSheets(selectedExcelFile)
	if err != nil {
		src.Logger(src.Error, "main.go/run()/src.GetSheets()", "Error when executing GetSheets()", "シート一覧の取得中にエラーが発生しました")
		return err
	}

	fmt.Printf("取得したシート一覧: %v\n", sheetList) // dev

	// 使用するシートを選択する
	var selectedSheet string
	if len(sheetList) > 1 {
		if selectedSheet, err = src.Select("使用するシートを選択してください", sheetList); err != nil {
			src.Logger(src.Error, "main.go/run()/src.Select()", "Error when executing Select()", "Select()の実行中にエラーが発生しました")
			return err
		}
	} else {
		selectedSheet = sheetList[0]
	}

	fmt.Printf("選択されたシート%v\n", selectedSheet) // dev

	// Excelデータの取得
	var excelData [][]string
	if excelData, err = src.ReadExcel(selectedExcelFile, selectedSheet); err != nil {
		src.Logger(src.Error, "main.go/run()/src.ReadExcel()", "Error when executing ReadExcel()", "ReadExcel()の実行中にエラーが発生しました")
		return err
	}

	fmt.Printf("取得したExcelデータ: %v\n", excelData) // dev

	// Excelデータを変換する

	return nil
}

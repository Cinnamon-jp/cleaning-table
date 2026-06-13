package main

import (
	"cleaning-table/src/excel"
	"cleaning-table/src/model"
	"cleaning-table/src/pdf"
	"cleaning-table/src/shuffle"
	"cleaning-table/src/util"
	"fmt"
)

func main() {
	if err := run(); err != nil {
		util.Logger(util.Error, "main.go/main()/run()", "Error when executing run()", "run()の実行中にエラーが発生しました")
	}
}

func run() error {
	// Excelファイルリストの取得
	var excelFiles []string
	var err error
	if excelFiles, err = excel.GetExcel("."); err != nil {
		util.Logger(util.Error, "main.go/run()/excel.GetExcel()", "Error when executing GetExcel()", "GetExcel()の実行中にエラーが発生しました")
		return err
	}

	fmt.Printf("カレントディレクトリにあるエクセルファイル一覧: %v\n", excelFiles) // dev

	// 使用するExcelファイルを選択する
	var selectedExcelFile string
	if len(excelFiles) > 1 {
		if selectedExcelFile, err = util.Select("使用するExcelファイルを選択してください", excelFiles); err != nil {
			util.Logger(util.Error, "main.go/run()/util.Select()", "Error when executing Select()", "Select()の実行中にエラーが発生しました")
			return err
		}
	} else {
		selectedExcelFile = excelFiles[0]
	}

	fmt.Printf("選択されたExcelファイル%v\n", selectedExcelFile) // dev

	// シートの一覧リストを取得
	sheetList, err := excel.GetSheets(selectedExcelFile)
	if err != nil {
		util.Logger(util.Error, "main.go/run()/excel.GetSheets()", "Error when executing GetSheets()", "シート一覧の取得中にエラーが発生しました")
		return err
	}

	fmt.Printf("取得したシート一覧: %v\n", sheetList) // dev

	// 使用するシートを選択する
	var selectedSheet string
	if len(sheetList) > 1 {
		if selectedSheet, err = util.Select("使用するシートを選択してください", sheetList); err != nil {
			util.Logger(util.Error, "main.go/run()/util.Select()", "Error when executing Select()", "Select()の実行中にエラーが発生しました")
			return err
		}
	} else {
		selectedSheet = sheetList[0]
	}

	fmt.Printf("選択されたシート%v\n", selectedSheet) // dev

	// Excelデータの取得
	var excelData [][]string
	if excelData, err = excel.ReadExcel(selectedExcelFile, selectedSheet); err != nil {
		util.Logger(util.Error, "main.go/run()/excel.ReadExcel()", "Error when executing ReadExcel()", "ReadExcel()の実行中にエラーが発生しました")
		return err
	}

	fmt.Printf("取得したExcelデータ: %v\n", excelData) // dev

	// Excelデータを変換する
	var convertedData []model.PostSet
	if convertedData, err = excel.ConvertExcel(excelData); err != nil {
		util.Logger(util.Error, "main.go/run()/excel.ConvertExcel()", "Error when executing ConvertExcel()", "ConvertExcel()の実行中にエラーが発生しました")
		return err
	}

	fmt.Printf("変換後のデータ: %v\n", convertedData) // dev

	// 役職をシャッフル
	var shuffledPostSet []model.ShuffledPostSet
	if shuffledPostSet, err = shuffle.SimpleShuffle(convertedData); err != nil {
		util.Logger(util.Error, "main.go/run()/shuffle.SimpleShuffle()", "Error when executing SimpleShuffle()", "SimpleShuffle()の実行中にエラーが発生しました")
		return err
	}

	fmt.Printf("シャッフル後のデータ: %v\n", shuffledPostSet) // dev

	// PDF出力
	if err = pdf.OutputPdf(shuffledPostSet); err != nil {
		util.Logger(util.Error, "main.go/run()/pdf.OutputPdf()", "Error when executing OutputPdf()", "OutputPdf()の実行中にエラーが発生しました")
		return err
	}

	return nil
}

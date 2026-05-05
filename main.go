package main

import (
	"cleaning-table/util"
	"os"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		util.Logger.Error("main.go: main()", err.Error(), "予期せぬエラーが発生しました")
		os.Exit(1)
	}
}

func run() error {
	
	excelFiles, err := findExcelFiles()
	if (err != nil) || (len(excelFiles) == 0) {
		util.Logger.Error(
			"main.go: findExcelFiles()",
			"Couldn't find excel files",
			"エクセルファイルを見つけることができませんでした",
		)
		return err
	}

	// 同じ階層にエクセルファイルが複数存在する場合は使用するファイルを選択させる
	excelFileToUse := ""
	if len(excelFiles) == 1 {
		excelFileToUse = excelFiles[0]
	} else {
		excelFileToUse, err = util.ChooseOne(
			"Choose the excel file you want to use. / 使用するエクセルファイルを選択してください。",
			excelFiles,
		)
		if err != nil {
			util.Logger.Error(
				"main.go: util.ChooseOne()",
				"Couldn't choose excel file",
				"エクセルファイルを選択できませんでした",
			)
			return err
		}
	}

	// エクセルファイルからデータを取得
	excelData, err := getExcelData(excelFileToUse)
	if err != nil {
		util.Logger.Error(
			"main.go: GetExcelData()",
			"Couldn't get excel data",
			"エクセルデータの内容を取得できませんでした",
		)
		return err
	}

	// 各列の部屋番号に対してタスクをランダムに割り振る
	assignments := assignTasks(*excelData)

	// 階ごとにデータを整形する
	floorAssignments := groupByFloor(assignments)

	// 今後の処理（PDF出力など）で floorAssignments を使用する想定
	_ = floorAssignments

	return nil
}

// findExcelFiles はカレントディレクトリから .xlsx ファイルを検索します。
func findExcelFiles() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".xlsx" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
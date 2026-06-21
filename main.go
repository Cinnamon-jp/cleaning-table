// Package main は掃除当番表アプリケーションのエントリーポイントを提供する
package main

import (
	"fmt"
	"log/slog"

	"cleaning-table/src/excel"
	"cleaning-table/src/history"
	"cleaning-table/src/model"
	"cleaning-table/src/pdf"
	"cleaning-table/src/shuffle"
	"cleaning-table/src/util"
)

const historyFileName = "cleaning_history.json"

func main() {
	if err := run(); err != nil {
		slog.Error("error occurred / エラーが発生しました", slog.Any("err", err))
	}
}

func run() error {
	// Excelファイルリストの取得
	var excelFiles []string
	var err error
	if excelFiles, err = excel.GetExcel("."); err != nil {
		return fmt.Errorf("getting Excel files / Excelファイル一覧の取得中: %w", err)
	}

	// 使用するExcelファイルを選択する
	var selectedExcelFile string
	if len(excelFiles) > 1 {
		if selectedExcelFile, err = util.Select("使用するExcelファイルを選択してください", excelFiles); err != nil {
			return fmt.Errorf("selecting the Excel file / Excelファイルの選択中: %w", err)
		}
	} else {
		selectedExcelFile = excelFiles[0]
	}

	// シートの一覧リストを取得
	sheetList, err := excel.GetSheets(selectedExcelFile)
	if err != nil {
		return fmt.Errorf("getting sheets list / シート一覧の取得中: %w", err)
	}

	// 使用するシートを選択する
	var selectedSheet string
	if len(sheetList) > 1 {
		if selectedSheet, err = util.Select("使用するシートを選択してください", sheetList); err != nil {
			return fmt.Errorf("selecting the sheet / シートの選択中: %w", err)
		}
	} else {
		selectedSheet = sheetList[0]
	}

	// Excelデータの取得
	var excelData [][]string
	if excelData, err = excel.ReadExcel(selectedExcelFile, selectedSheet); err != nil {
		return fmt.Errorf("reading the Excel data / Excelデータの読み込み中: %w", err)
	}

	// Excelデータを変換する
	var convertedData []model.PostSet
	if convertedData, err = excel.ConvertExcel(excelData); err != nil {
		return fmt.Errorf("converting the Excel data / Excelデータの変換中: %w", err)
	}

	// 履歴ファイルの読み込み
	hist, err := history.LoadHistory(historyFileName)
	if err != nil {
		return fmt.Errorf("loading history / 履歴の読み込み中: %w", err)
	}

	// フィンガープリントを生成し、Excel構成の変更を検知
	fingerprint := history.GenerateFingerprint(convertedData)
	history.CheckAndResetHistory(hist, fingerprint)

	// 履歴からカウントマップを構築
	countMap := history.GetCountMap(hist)

	// 役職を重み付きシャッフル
	var shuffledPostSet []model.ShuffledPostSet
	if shuffledPostSet, err = shuffle.EvenShuffle(convertedData, countMap); err != nil {
		return fmt.Errorf("shuffling the post sets / 役職セットのシャッフル中: %w", err)
	}

	// 履歴の累計カウントを更新して保存
	history.UpdateCounts(hist, shuffledPostSet)
	if err = history.SaveHistory(historyFileName, hist); err != nil {
		return fmt.Errorf("saving history / 履歴の保存中: %w", err)
	}

	// PDF出力
	if err = pdf.OutputPdf(shuffledPostSet); err != nil {
		return fmt.Errorf("outputting the PDF / PDF出力中: %w", err)
	}

	return nil
}

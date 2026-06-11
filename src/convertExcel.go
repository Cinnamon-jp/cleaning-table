// Package src はプログラムを構成する部品に分割して格納する
package src

import (
	"errors"
	"strconv"
	"strings"
)

// ConvertExcel はExcelから取り出したデータを処理しやすいように変換する
func ConvertExcel(excelData [][]string) (convertedData [][]string, err error) {
	// Excelデータの文法をチェックする
	var isOK bool
	if isOK, err = checkExcel(excelData); !isOK || err != nil {
		Logger(Error, "convertExcel.go/ConvertExcel()/checkExcel()", "Error when executing checkExcel()", "checkExcel()の実行中にエラーが発生しました")
		return nil, err
	}
	
	// コメント列を削除
	noCommentExcelData := removeComment(excelData)

	// 部屋番号範囲の展開


	return nil, nil
}

// checkExcel はExcelの文法チェックを行う 未完成
func checkExcel(excelData [][]string) (isOK bool, err error) {
	return true, nil
}

// removeComment はコメント行(1行目+1列目)を削除する
func removeComment(excelData [][]string) [][]string {
	// データが空、または1行しかない場合は空のスライスを返す
	if len(excelData) <= 1 {
		return [][]string{}
	}

	// 元のスライスに影響を与えないよう、新しいスライスを作成する
	var result [][]string
	// 1行目をスキップ (excelData[1:])
	for _, row := range excelData[1:] {
		// 列が1列以下しかない場合は空の行を追加
		if len(row) <= 1 {
			result = append(result, []string{})
		} else {
			// 1列目をスキップした要素をコピーして新しい行を作成
			newRow := make([]string, len(row)-1)
			copy(newRow, row[1:])
			result = append(result, newRow)
		}
	}

	return result
}

// unfoldRoomNumber は文字列を受け取って部屋番号を分割し、範囲指定された部屋番号を展開してintのスライスを返す
func unfoldRoomNumber(s string) ([]int, error) {
	var result []int

	// 空文字の場合は空のスライスを返す
	if strings.TrimSpace(s) == "" {
		return result, nil
	}

	// カンマで分割
	parts := strings.Split(s, ",")

	for _, part := range parts {
		// 前後の空白を削除
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// ":"が含まれているか確認（範囲指定）
		if strings.Contains(part, ":") {
			rangeParts := strings.Split(part, ":")
			if len(rangeParts) == 2 {
				startStr := strings.TrimSpace(rangeParts[0])
				endStr := strings.TrimSpace(rangeParts[1])

				start, err1 := strconv.Atoi(startStr)
				end, err2 := strconv.Atoi(endStr)

				// パースに成功し、start <= end の場合のみ展開
				if err1 == nil && err2 == nil && start <= end {
					for i := start; i <= end; i++ {
						result = append(result, i)
					}
				} else {
					Logger(Error, "convertExcel.go/unfoldRoomNumber()/err1 == nil && err2 == nil && start <= end", "Invalid range format in a Excel file", "Excelの範囲指定文法が不正です")
					return nil, errors.New("Invalid range format in a Excel file")
				}
			} else {
				Logger(Error, "convertExcel.go/unfoldRoomNumber()/len(rangeParts) == 2", "Invalid range format in a Excel file", "Excelの範囲指定文法が不正です")
				return  nil, errors.New("Invalid range format in a Excel file")
			}
		} else {
			// 単一の部屋番号
			num, err := strconv.Atoi(part)
			if err == nil {
				result = append(result, num)
			}
		}
	}

	return result, nil
}
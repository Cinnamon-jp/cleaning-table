// Package excel はExcelファイルの読み込み・変換処理を格納する
package excel

import (
	"cleaning-table/src/model"
	"cleaning-table/src/util"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ConvertExcel はExcelから取り出したデータを処理しやすいように変換する
func ConvertExcel(excelData [][]string) (convertedData []model.PostSet, err error) {
	// Excelデータの文法をチェックする
	var isOK bool
	if isOK, err = checkExcel(excelData); !isOK || err != nil {
		util.Logger(util.Error, "convertExcel.go/ConvertExcel()/checkExcel()", "Error when executing checkExcel()", "checkExcel()の実行中にエラーが発生しました")
		return nil, err
	}

	// コメント列を削除
	noCommentExcelData := removeComment(excelData)

	var postSets []model.PostSet

	for _, row := range noCommentExcelData {
		// 部屋番号範囲の展開
		unfoldedRoomNumber, err := unfoldRoomNumber(row[0])
		if err != nil {
			util.Logger(util.Error, "convertExcel.go/ConvertExcel()/unfoldRoomNumber()", "Error when executing unfoldRoomNumber()", "unfoldRoomNumber()の実行中にエラーが発生しました")
			return nil, err
		}

		// 役職の名前と数の分解
		unfoldedPosts, unfoldedPostCounts, questionIndex, err := unfoldPost(row[1:])
		if err != nil {
			util.Logger(util.Error, "convertExcel.go/ConvertExcel()/unfoldPost()", "Error when executing unfoldPost()", "unfoldPost()の実行中にエラーが発生しました")
			return nil, err
		}

		// `?` が含まれていた場合、RoomNumbersの個数から他の役職数の合計を引いて算出する
		if questionIndex >= 0 {
			totalRooms := len(unfoldedRoomNumber)
			var sumOtherCounts int
			for i, count := range unfoldedPostCounts {
				if i != questionIndex {
					sumOtherCounts += count
				}
			}
			remainingCount := totalRooms - sumOtherCounts
			if remainingCount < 0 {
				util.Logger(util.Error, "convertExcel.go/ConvertExcel()/questionIndex", "Calculated remaining count is negative", "?の役職数の計算結果が負の値になりました")
				return nil, fmt.Errorf("calculated remaining count is negative: totalRooms=%d, sumOtherCounts=%d", totalRooms, sumOtherCounts)
			}
			unfoldedPostCounts[questionIndex] = remainingCount
		}

		postSets = append(postSets, model.PostSet{
			RoomNumbers: unfoldedRoomNumber,
			Posts:       unfoldedPosts,
			PostCounts:  unfoldedPostCounts,
		})
	}

	return postSets, nil
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
		// 列が1列以下しかない場合はスキップ
		if len(row) <= 1 {
			continue
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

	// 空文字の場合は空文字列を返す
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
					util.Logger(util.Warn, "convertExcel.go/unfoldRoomNumber()/err1 == nil && err2 == nil && start <= end", "Invalid range format in a Excel file", "Excelの範囲指定文法が不正です")
					return nil, errors.New("invalid range format in a Excel file")
				}
			} else {
				util.Logger(util.Warn, "convertExcel.go/unfoldRoomNumber()/len(rangeParts) == 2", "Invalid range format in a Excel file", "Excelの範囲指定文法が不正です")
				return nil, errors.New("invalid range format in a Excel file")
			}
		} else {
			// 単一の部屋番号
			num, err := strconv.Atoi(part)
			if err != nil {
				util.Logger(util.Warn, "convertExcel.go/unfoldRoomNumber()/strconv.Atoi", "Invalid room number format in a Excel file", "Excelの部屋番号の文法が不正です")
				return nil, errors.New("invalid room number format in a Excel file")
			}
			result = append(result, num)
		}
	}

	return result, nil
}

// unfoldPost は []役職名*役職数 を受け取って []役職名 と []役職数 を返す
// 役職数に "?" が含まれている場合、該当インデックスを questionIndex として返す(存在しない場合は -1)
func unfoldPost(postsAndCounts []string) (posts []string, postCounts []int, questionIndex int, err error) {
	var returnPosts []string
	var returnPostCounts []int
	questionIndex = -1

	for _, postAndCount := range postsAndCounts {
		// 不要な空白を削除
		postAndCount = strings.TrimSpace(postAndCount)

		// 空文字はスキップ
		if postAndCount == "" {
			continue
		}

		// *で2つに分割
		lastAsteriskIndex := strings.LastIndex(postAndCount, "*")
		var post string
		var postCountStr string

		// * が含まれていない場合は人数を1とみなす
		if lastAsteriskIndex == -1 {
			post = postAndCount
			postCountStr = "1"
		} else {
			post = strings.TrimSpace(postAndCount[:lastAsteriskIndex])
			postCountStr = strings.TrimSpace(postAndCount[lastAsteriskIndex+1:])
		}

		// どちらかが空文字の場合はエラー
		if post == "" || postCountStr == "" {
			util.Logger(util.Warn, "convertExcel.go/unfoldPost/post == \"\" || postCountStr == \"\"", "Invalid post format in a Excel file", "Excelの役職名の文法が不正です")
			return nil, nil, -1, errors.New("invalid post format in a Excel file")
		}

		// 役職名をスライスに追加
		returnPosts = append(returnPosts, post)

		// 役職数が "?" の場合は仮に0を入れて、インデックスを記録する
		if postCountStr == "?" {
			questionIndex = len(returnPostCounts)
			returnPostCounts = append(returnPostCounts, 0)
			continue
		}

		// 役職数をスライスに追加
		returnPostCount, err := strconv.Atoi(postCountStr)
		if err != nil {
			util.Logger(util.Error, "convertExcel.go/unfoldPost/strconv.Atoi", "Invalid post count format in a Excel file", "Excelの役職数の文法が不正です")
			return nil, nil, -1, fmt.Errorf("invalid post count format in a Excel file: %w", err)
		}
		returnPostCounts = append(returnPostCounts, returnPostCount)
	}

	return returnPosts, returnPostCounts, questionIndex, nil
}

// Package src は掃除当番表の生成やExcelファイルのパースなど、アプリケーションの主要なロジックを提供します。
package src

import (
	"cleaning-table/src/util"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type UnfoldedExcelData struct {
	RoomNumbers [][]int
	Tasks       [][]string
}

// GetExcelData は指定されたパスのエクセルファイルを読み込み、内容を2次元スライスとして返します。
func GetExcelData(path string) (*UnfoldedExcelData, error) {
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

	unfolded, err := unfoldExcelData(excel)
	if err != nil {
		util.Logger.Error(
			"excel.go: unfoldExcelData()",
			"Couldn't unfold excel data",
			"エクセルデータの展開に失敗しました",
		)
		return nil, err
	}
	return &unfolded, nil
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
		if closeErr := f.Close(); closeErr != nil {
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

	var roomNumbers [][]int
	var tasks [][]string

	numCols := len(roomNumberStrings)
	for col := 0; col < numCols; col++ {
		// 部屋番号の展開
		rooms, err := parseRoomNumbers(roomNumberStrings[col])
		if err != nil {
			util.Logger.Error(
				"excel.go: unfoldExcelData()",
				fmt.Sprintf("Failed to parse room numbers at column %d: %v", col+1, err),
				fmt.Sprintf("%d列目の部屋番号の解析に失敗しました", col+1),
			)
			return UnfoldedExcelData{}, err
		}
		roomNumbers = append(roomNumbers, rooms)

		// 該当列のタスク文字列を収集
		var colTasks []string
		for row := 0; row < len(taskStrings); row++ {
			if col < len(taskStrings[row]) {
				colTasks = append(colTasks, taskStrings[row][col])
			}
		}

		// タスクの展開
		parsedTasks, err := parseTasks(colTasks, len(rooms))
		if err != nil {
			util.Logger.Error(
				"excel.go: unfoldExcelData()",
				fmt.Sprintf("Failed to parse tasks at column %d: %v", col+1, err),
				fmt.Sprintf("%d列目の役職の解析に失敗しました", col+1),
			)
			return UnfoldedExcelData{}, err
		}
		tasks = append(tasks, parsedTasks)
	}

	return UnfoldedExcelData{
		RoomNumbers: roomNumbers,
		Tasks:       tasks,
	}, nil
}

// parseRoomNumbers は部屋番号の文字列（例："101, 103:105"）を解析し、整数のスライスとして返します。
func parseRoomNumbers(s string) ([]int, error) {
	var result []int
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, ":") {
			rangeParts := strings.Split(part, ":")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("failed to parse start of range %s: %w", part, err)
			}
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("failed to parse end of range %s: %w", part, err)
			}
			if start > end {
				return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
			}
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("failed to parse room number %s: %w", part, err)
			}
			result = append(result, num)
		}
	}
	return result, nil
}

// parseTasks はタスクの文字列スライス（例：["フロア*4", "自室清掃*?"]）を解析し、
// totalRooms と数が一致するように展開した文字列スライスを返します。
func parseTasks(taskStrs []string, totalRooms int) ([]string, error) {
	var result []string
	type taskDef struct {
		name   string
		count  int
		isAuto bool
	}
	var defs []taskDef
	fixedTotal := 0
	autoIndex := -1

	for _, s := range taskStrs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		idx := strings.LastIndex(s, "*")
		if idx == -1 {
			return nil, fmt.Errorf("invalid task format (missing '*'): %s", s)
		}

		name := strings.TrimSpace(s[:idx])
		countStr := strings.TrimSpace(s[idx+1:])

		if countStr == "?" {
			if autoIndex != -1 {
				return nil, errors.New("multiple auto assignments (?) found")
			}
			defs = append(defs, taskDef{name: name, count: 0, isAuto: true})
			autoIndex = len(defs) - 1
		} else {
			count, err := strconv.Atoi(countStr)
			if err != nil {
				return nil, fmt.Errorf("invalid count format in %s: %w", s, err)
			}
			defs = append(defs, taskDef{name: name, count: count, isAuto: false})
			fixedTotal += count
		}
	}

	if autoIndex != -1 {
		autoCount := totalRooms - fixedTotal
		if autoCount < 0 {
			return nil, fmt.Errorf("total fixed tasks (%d) exceeds total rooms (%d)", fixedTotal, totalRooms)
		}
		defs[autoIndex].count = autoCount
	} else {
		if fixedTotal != totalRooms {
			return nil, fmt.Errorf("total tasks (%d) does not match total rooms (%d)", fixedTotal, totalRooms)
		}
	}

	for _, d := range defs {
		for i := 0; i < d.count; i++ {
			result = append(result, d.name)
		}
	}

	return result, nil
}

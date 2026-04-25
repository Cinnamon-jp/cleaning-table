package main

import (
	"errors"
	"log/slog"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func checkExcelSyntax(excel [][]string) error {
	for i, row := range excel {
		for _, cell := range row {
			// 1行目の部屋番号をチェック
			if i == 0 {
				if _, err := roomStringToNumbers(cell); err != nil {
					return err
				}
			// 2行目以降の清掃役職をチェック
			} else {
			}
		}
	}
	return nil
}

// 正規表現
var (
	roomNumberRegex = regexp.MustCompile(`^\d+(~\d+)?$`)
	existRoomRegex  = regexp.MustCompile(VALID_ROOM_NUMBER)
	taskRegex       = regexp.MustCompile(`^.+\*(\d+|\?)$`)
)

// roomStringToNumbersは部屋番号文字列を解析し、intのスライスを返す
func roomStringToNumbers(str string) ([]int, error) {
	// 文字列を正規化
	str = strings.TrimSpace(str)
    roomNumbers := strings.Split(str, ",")

	// 戻り値の初期化
	returnSlice := []int{}

	// 各要素をチェック
	for i, roomNumber := range roomNumbers {
		// エラーメッセージ用のセル名を代入
		cellName, err := indexToCell(0, i)
		if err != nil {
			return []int{}, err
		}

		// 正規表現で部屋番号のフォーマットをチェック
		if isValidRoomNumber := roomNumberRegex.MatchString(roomNumber); !isValidRoomNumber {
			slog.Error("Invalid room number", "room", roomNumber, "cell", cellName)
			return []int{}, errors.New("Invalid room number syntax: Room number range is descending (部屋番号の範囲が降順になっています)")
		}

		// 部屋番号範囲指定を処理
		if roomNumbers := strings.Split(roomNumber, "~"); len(roomNumbers) == 2 {
			// 整数に変換
			startNumber, err := strconv.Atoi(roomNumbers[0])
			if err != nil {
				return []int{}, err
			}
			endNumber, err := strconv.Atoi(roomNumbers[1])
			if err != nil {
				return []int{}, err
			}
			// 範囲が昇順になっているかチェック
			if startNumber > endNumber {
				slog.Error("Invalid room number syntax: Room number range is descending", "cell", cellName)
				return []int{}, errors.New("Invalid room number syntax")
			}
			// startNumberからendNumberまでの部屋番号を追加
			for i := startNumber; i <= endNumber; i++ {
				returnSlice = append(returnSlice, i)
			}

		// 単部屋番号指定を処理
		} else {
			// 整数に変換
			roomNumber, err := strconv.Atoi(roomNumber)
			if err != nil {
				return []int{}, err
			}
			// 部屋番号を追加
			returnSlice = append(returnSlice, roomNumber)
		}
	}

	// 重複した部屋番号がないかチェック
	seen := make(map[int]bool)
	for _, roomNumber := range returnSlice {
		if seen[roomNumber] {
			slog.Error("Invalid room number syntax: Duplicate room number (部屋番号が重複しています)")
			return []int{}, errors.New("Invalid room number syntax")
		}
		seen[roomNumber] = true
	}

	// 昇順に並び替え
	sort.Ints(returnSlice)

	return returnSlice, nil
}

// taskStringToSliceは役役職文字列解析して役職数分のスライスを返す
func taskStringToSlice(tasks []string) ([]string, error) {
}

// indexToCellは0-indexed [row, col] pairをExcelセル名に変換する(例:"A1")
func indexToCell(row, col int) (string, error) {
	if row < 0 || row > 1_048_575 || col < 0 || col > 16_383 {
		return "", errors.New("index out of range")
	}

	var colStr string
	c := col
	for c >= 0 {
		remainder := c % 26
		colStr = string(rune('A'+remainder)) + colStr
		c = (c / 26) - 1
	}

	rowStr := strconv.Itoa(row + 1)
	return colStr + rowStr, nil
}

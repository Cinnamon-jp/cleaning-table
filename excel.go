package main

import (
	"errors"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

func checkExcelSyntax(excel [][]string) error {
	for i, row := range excel {
		for j, cell := range row {
			// 1行目の部屋番号をチェック
			if i == 0 {
				roomNumbers := strings.Split(cell, ",")
				for _, roomNumber := range roomNumbers {
					if !isRoomNumber(roomNumber) {
						cellName, err := indexToCell(i, j)
						if err != nil {
							return err
						}
						slog.Error("Invalid room number: %s in %s", cell, cellName)
					}
				}
			// 2行目以降の清掃役職をチェック
			} else {
				if !isTask(cell) {
					cellName, err := indexToCell(i, j)
					if err != nil {
						return err
					}
					slog.Error("Invalid task: %s in %s", cell, cellName)
				}
			}
		}
	}
	return nil
}

// 正規表現
var (
	roomNumberRegexp = regexp.MustCompile(`^\d{3}(~\d{3})?$`)
	taskRegexp       = regexp.MustCompile(`^.+\*(\d+|\?)$`)
)

// isRoomNumberは`,`で区切った要素をチェックする
func isRoomNumber(str string) bool {
	return roomNumberRegexp.MatchString(str)
}

// isTaskは1行目以外の要素をチェックする
func isTask(str string) bool {
	return taskRegexp.MatchString(str)
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

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Role は役職名とその必要な割り当て人数を保持します。
type Role struct {
	Name  string
	Count int
}

// ColumnData は特定の列について解析されたすべての部屋番号と役職を保持します。
type ColumnData struct {
	Rooms []int
	Roles []Role
}

// ExcelParser はエクセルファイルの構造を解析および検証します。
type ExcelParser struct {
	spaceRegex *regexp.Regexp
}

func NewExcelParser() *ExcelParser {
	return &ExcelParser{
		spaceRegex: regexp.MustCompile(`\d\s+\d`),
	}
}

// ValidateSpaces は数字の間に無効なスペースがあるかどうかをチェックします。
func (p *ExcelParser) ValidateSpaces(cellValue string) error {
	if p.spaceRegex.MatchString(cellValue) {
		return fmt.Errorf("invalid space between numbers: %q", cellValue)
	}
	return nil
}

// ParseRooms は部屋番号の文字列（例："101, 103~105"）を部屋番号のスライスに解析します。
func (p *ExcelParser) ParseRooms(cell string) ([]int, error) {
	if err := p.ValidateSpaces(cell); err != nil {
		return nil, err
	}

	cell = strings.ReplaceAll(cell, " ", "")
	if cell == "" {
		return nil, nil
	}

	parts := strings.Split(cell, ",")
	var rooms []int
	seen := make(map[int]bool)

	for _, part := range parts {
		if strings.Contains(part, "~") {
			rangeParts := strings.Split(part, "~")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %q", part)
			}

			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid start room number: %q", rangeParts[0])
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid end room number: %q", rangeParts[1])
			}

			if start >= end {
				return nil, fmt.Errorf("range must be ascending: %q", part)
			}

			for r := start; r <= end; r++ {
				// ユーザーが 149~201 のように指定した場合に 150-199 のような無効な部屋をスキップ、またはエラーにする。
				// 仕様上「指定するときは必ず昇順」となっており、有効な部屋は 1F~9F、各階 01~49。
				// 149~201 と指定された場合 150 などが含まれるため、すべての生成された部屋を検証する。
				if err := validateRoom(r); err != nil {
					return nil, err
				}
				if seen[r] {
					return nil, fmt.Errorf("duplicate room number: %d", r)
				}
				seen[r] = true
				rooms = append(rooms, r)
			}
		} else {
			r, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid room number: %q", part)
			}
			if err := validateRoom(r); err != nil {
				return nil, err
			}
			if seen[r] {
				return nil, fmt.Errorf("duplicate room number: %d", r)
			}
			seen[r] = true
			rooms = append(rooms, r)
		}
	}
	return rooms, nil
}

func validateRoom(r int) error {
	floor := r / 100
	room := r % 100
	if floor < 1 || floor > 9 {
		return fmt.Errorf("invalid floor in room %d", r)
	}
	if room < 1 || room > 49 {
		return fmt.Errorf("invalid room number %d, must be between 01 and 49", r)
	}
	return nil
}

// ParseRole は役職定義セル（例："ゴミ分別*3" または "フロア*?"）を解析します。
// 役職名、人数（'?' の場合は -1）、およびエラーを返します。
func (p *ExcelParser) ParseRole(cell string) (Role, error) {
	if err := p.ValidateSpaces(cell); err != nil {
		return Role{}, err
	}

	cell = strings.ReplaceAll(cell, " ", "")
	if cell == "" {
		return Role{}, nil // 空のセル
	}

	parts := strings.Split(cell, "*")
	if len(parts) != 2 {
		return Role{}, fmt.Errorf("invalid role format: %q", cell)
	}

	roleName := parts[0]
	if roleName == "" {
		return Role{}, fmt.Errorf("role name cannot be empty: %q", cell)
	}

	countStr := parts[1]
	if countStr == "?" {
		return Role{Name: roleName, Count: -1}, nil
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return Role{}, fmt.Errorf("invalid role count: %q in %q", countStr, cell)
	}
	if count <= 0 {
		return Role{}, fmt.Errorf("role count must be positive: %d in %q", count, cell)
	}

	return Role{Name: roleName, Count: count}, nil
}

// ValidateColumn は列のデータの整合性をチェックします。
// '?' の人数を解決し、重複する役職名がないか確認し、人数の合計が一致するか検証します。
func (p *ExcelParser) ValidateColumn(colData *ColumnData) error {
	totalRooms := len(colData.Rooms)
	if totalRooms == 0 {
		// 部屋がない場合、エラーにするか単に nil を返してスキップする。
		// 通常、部屋がない列に役職が指定されているべきではないためエラーとする。
		if len(colData.Roles) > 0 {
			return fmt.Errorf("roles specified but no rooms found")
		}
		return nil
	}

	seenRoles := make(map[string]bool)
	totalRoles := 0
	questionMarkIdx := -1

	for i, role := range colData.Roles {
		if seenRoles[role.Name] {
			return fmt.Errorf("duplicate role name: %q", role.Name)
		}
		seenRoles[role.Name] = true

		if role.Count == -1 {
			if questionMarkIdx != -1 {
				return fmt.Errorf("multiple '?' used in the same column")
			}
			questionMarkIdx = i
		} else {
			totalRoles += role.Count
		}
	}

	if questionMarkIdx != -1 {
		remaining := totalRooms - totalRoles
		if remaining < 0 {
			return fmt.Errorf("total roles (%d) exceeds total rooms (%d)", totalRoles, totalRooms)
		}
		colData.Roles[questionMarkIdx].Count = remaining
	} else {
		if totalRoles != totalRooms {
			return fmt.Errorf("total roles (%d) does not match total rooms (%d)", totalRoles, totalRooms)
		}
	}

	return nil
}

// ExcelData はエクセルファイルから解析および検証されたデータを表します。
type ExcelData struct {
	Columns []ColumnData
}

// ParseExcelFile はエクセルファイルを読み込み、列ごとにデータを解析し、ExcelData を返します。
func (p *ExcelParser) ParseExcelFile(filePath string) (*ExcelData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel file has no sheets")
	}

	sheetName := sheets[0] // 最初のシートを想定
	cols, err := f.GetCols(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns from sheet: %w", err)
	}

	var excelData ExcelData

	for colIdx, col := range cols {
		if len(col) == 0 {
			continue // 空の列
		}

		colName, err := excelize.ColumnNumberToName(colIdx + 1)
		if err != nil {
			return nil, fmt.Errorf("failed to get column name: %w", err)
		}

		// セル A1 は 1行目
		cellName := fmt.Sprintf("%s1", colName)
		roomStr := col[0]
		if strings.TrimSpace(roomStr) == "" {
			continue // 1行目に部屋がない場合、列をスキップする
		}

		rooms, err := p.ParseRooms(roomStr)
		if err != nil {
			return nil, fmt.Errorf("cell %s: %v", cellName, err)
		}

		colData := ColumnData{
			Rooms: rooms,
		}

		// 2行目以降（インデックス 1 以降）
		for rowIdx := 1; rowIdx < len(col); rowIdx++ {
			roleStr := col[rowIdx]
			if strings.TrimSpace(roleStr) == "" {
				continue
			}

			cellName := fmt.Sprintf("%s%d", colName, rowIdx+1)
			role, err := p.ParseRole(roleStr)
			if err != nil {
				return nil, fmt.Errorf("cell %s: %v", cellName, err)
			}
			colData.Roles = append(colData.Roles, role)
		}

		if err := p.ValidateColumn(&colData); err != nil {
			return nil, fmt.Errorf("column %s error: %v", colName, err)
		}

		excelData.Columns = append(excelData.Columns, colData)
	}

	return &excelData, nil
}

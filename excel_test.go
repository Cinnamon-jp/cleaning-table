package main

import (
	"os"
	"reflect"
	"testing"
)

func TestParseRooms(t *testing.T) {
	parser := NewExcelParser()

	tests := []struct {
		name      string
		input     string
		want      []int
		expectErr bool
	}{
		{
			name:      "Valid single room",
			input:     "101",
			want:      []int{101},
			expectErr: false,
		},
		{
			name:      "Valid multiple rooms",
			input:     "101, 102",
			want:      []int{101, 102},
			expectErr: false,
		},
		{
			name:      "Valid range",
			input:     "101~103",
			want:      []int{101, 102, 103},
			expectErr: false,
		},
		{
			name:      "Valid mix of single and range",
			input:     "101, 103~105, 107",
			want:      []int{101, 103, 104, 105, 107},
			expectErr: false,
		},
		{
			name:      "Valid multi floor",
			input:     "148~202",
			want:      []int{148, 149, 201, 202},
			expectErr: false,
		},
		{
			name:      "Valid only space",
			input:     " ",
			want:      nil,
			expectErr: false,
		},
		{
			name:      "Invalid spaces between numbers",
			input:     "10 1",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid Duplicate rooms",
			input:     "101, 101",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid Duplicate rooms in range",
			input:     "101, 101~102",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid Range not ascending",
			input:     "103~101",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid 3 numbers",
			input:     "101~102~103",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid characters",
			input:     "abc",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid start characters",
			input:     "abc~102",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid end characters",
			input:     "101~abc",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid room number (too high)",
			input:     "150",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid room number (too low)",
			input:     "099",
			want:      nil,
			expectErr: true,
		},
		{
			name:      "Invalid floor",
			input:     "1001",
			want:      nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseRooms(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseRooms() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRooms() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRole(t *testing.T) {
	parser := NewExcelParser()

	tests := []struct {
		name      string
		input     string
		want      Role
		expectErr bool
	}{
		{
			name:      "Valid role with count",
			input:     "ゴミ分別*3",
			want:      Role{Name: "ゴミ分別", Count: 3},
			expectErr: false,
		},
		{
			name:      "Valid role with ?",
			input:     "フロア*?",
			want:      Role{Name: "フロア", Count: -1},
			expectErr: false,
		},
		{
			name:      "Valid role with spaces",
			input:     "ゴ ミ 分 別 * 3",
			want:      Role{Name: "ゴミ分別", Count: 3},
			expectErr: false,
		},
		{
			name:      "Invalid role count",
			input:     "ゴミ分別*abc",
			want:      Role{},
			expectErr: true,
		},
		{
			name:      "Invalid format no asterisk",
			input:     "ゴミ分別3",
			want:      Role{},
			expectErr: true,
		},
		{
			name:      "Invalid count zero",
			input:     "ゴミ分別*0",
			want:      Role{},
			expectErr: true,
		},
		{
			name:      "Invalid count negative",
			input:     "ゴミ分別*-1",
			want:      Role{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseRole(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseRole() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateColumn(t *testing.T) {
	parser := NewExcelParser()

	tests := []struct {
		name      string
		input     *ColumnData
		wantRoles []Role
		expectErr bool
	}{
		{
			name: "Valid exact match",
			input: &ColumnData{
				Rooms: []int{101, 102, 103},
				Roles: []Role{
					{Name: "A", Count: 2},
					{Name: "B", Count: 1},
				},
			},
			wantRoles: []Role{
				{Name: "A", Count: 2},
				{Name: "B", Count: 1},
			},
			expectErr: false,
		},
		{
			name: "Valid with question mark",
			input: &ColumnData{
				Rooms: []int{101, 102, 103, 104, 105},
				Roles: []Role{
					{Name: "A", Count: 2},
					{Name: "B", Count: -1},
				},
			},
			wantRoles: []Role{
				{Name: "A", Count: 2},
				{Name: "B", Count: 3},
			},
			expectErr: false,
		},
		{
			name: "Mismatch sum",
			input: &ColumnData{
				Rooms: []int{101, 102, 103},
				Roles: []Role{
					{Name: "A", Count: 2},
					{Name: "B", Count: 2},
				},
			},
			wantRoles: nil,
			expectErr: true,
		},
		{
			name: "Negative remaining for question mark",
			input: &ColumnData{
				Rooms: []int{101, 102},
				Roles: []Role{
					{Name: "A", Count: 3},
					{Name: "B", Count: -1},
				},
			},
			wantRoles: nil,
			expectErr: true,
		},
		{
			name: "Multiple question marks",
			input: &ColumnData{
				Rooms: []int{101, 102, 103, 104},
				Roles: []Role{
					{Name: "A", Count: -1},
					{Name: "B", Count: -1},
				},
			},
			wantRoles: nil,
			expectErr: true,
		},
		{
			name: "Duplicate roles",
			input: &ColumnData{
				Rooms: []int{101, 102, 103},
				Roles: []Role{
					{Name: "A", Count: 2},
					{Name: "A", Count: 1},
				},
			},
			wantRoles: nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateColumn(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateColumn() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr {
				if !reflect.DeepEqual(tt.input.Roles, tt.wantRoles) {
					t.Errorf("ValidateColumn() transformed roles = %v, want %v", tt.input.Roles, tt.wantRoles)
				}
			}
		})
	}
}

func TestParseExcelFile(t *testing.T) {
	if _, err := os.Stat("test.xlsx"); os.IsNotExist(err) {
		t.Skip("test.xlsx does not exist, skipping test")
	}

	parser := NewExcelParser()
	data, err := parser.ParseExcelFile("test.xlsx")
	if err != nil {
		t.Fatalf("ParseExcelFile failed for test.xlsx: %v", err)
	}

	if data == nil || len(data.Columns) == 0 {
		t.Fatalf("ParseExcelFile returned empty data for test.xlsx")
	}

	// データがパースされたかどうかの基本的な確認
	t.Logf("Parsed %d columns from test.xlsx", len(data.Columns))
}

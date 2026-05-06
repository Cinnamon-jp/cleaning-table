package main

import (
	"reflect"
	"testing"
)

func TestDeleteCommentRow(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]string
		expected [][]string
	}{
		{
			name: "コメント行が含まれる場合",
			input: [][]string{
				{"#", "comment row"},
				{"101, 102", "task*1"},
				{"#", "another comment"},
				{"103", "task*2"},
			},
			expected: [][]string{
				{"101, 102", "task*1"},
				{"103", "task*2"},
			},
		},
		{
			name: "コメント行が含まれない場合",
			input: [][]string{
				{"101, 102", "task*1"},
				{"103", "task*2"},
			},
			expected: [][]string{
				{"101, 102", "task*1"},
				{"103", "task*2"},
			},
		},
		{
			name: "1列目が '#' ではないが '#' を含む場合 (削除されない)",
			input: [][]string{
				{"101", "#not a comment"},
			},
			expected: [][]string{
				{"101", "#not a comment"},
			},
		},
		{
			name:     "空のデータ",
			input:    [][]string{},
			expected: nil,
		},
		{
			name: "空の行が含まれる場合",
			input: [][]string{
				{},
				{"101", "task*1"},
			},
			expected: [][]string{
				{},
				{"101", "task*1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deleteCommentRow(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("deleteCommentRow() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseRoomNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
		wantErr  bool
	}{
		{
			name:     "カンマとコロンの複合",
			input:    "101, 103:105",
			expected: []int{101, 103, 104, 105},
			wantErr:  false,
		},
		{
			name:     "複数の範囲",
			input:    "202, 204:207, 209:211",
			expected: []int{202, 204, 205, 206, 207, 209, 210, 211},
			wantErr:  false,
		},
		{
			name:     "単一の部屋",
			input:    "301",
			expected: []int{301},
			wantErr:  false,
		},
		{
			name:    "不正なフォーマット",
			input:   "101:102:103",
			wantErr: true,
		},
		{
			name:    "無効な数値",
			input:   "10a",
			wantErr: true,
		},
		{
			name:    "逆転した範囲",
			input:   "105:103",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRoomNumbers(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRoomNumbers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseRoomNumbers() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseTasks(t *testing.T) {
	tests := []struct {
		name       string
		input      []string
		totalRooms int
		expected   []string
		wantErr    bool
	}{
		{
			name:       "固定数のタスク",
			input:      []string{"フロア*4"},
			totalRooms: 4,
			expected:   []string{"フロア", "フロア", "フロア", "フロア"},
			wantErr:    false,
		},
		{
			name:       "自動算出のタスク",
			input:      []string{"洗濯室*2", "自室清掃*?"},
			totalRooms: 8,
			expected:   []string{"洗濯室", "洗濯室", "自室清掃", "自室清掃", "自室清掃", "自室清掃", "自室清掃", "自室清掃"},
			wantErr:    false,
		},
		{
			name:       "全て自動算出",
			input:      []string{"全員*?"},
			totalRooms: 3,
			expected:   []string{"全員", "全員", "全員"},
			wantErr:    false,
		},
		{
			name:       "空文字無視",
			input:      []string{"フロア*2", "", "トイレ*1"},
			totalRooms: 3,
			expected:   []string{"フロア", "フロア", "トイレ"},
			wantErr:    false,
		},
		{
			name:       "不正なフォーマット(*なし)",
			input:      []string{"フロア4"},
			totalRooms: 4,
			wantErr:    true,
		},
		{
			name:       "複数の自動算出",
			input:      []string{"タスクA*?", "タスクB*?"},
			totalRooms: 4,
			wantErr:    true,
		},
		{
			name:       "人数超過",
			input:      []string{"フロア*5"},
			totalRooms: 4,
			wantErr:    true,
		},
		{
			name:       "人数不足",
			input:      []string{"フロア*3"},
			totalRooms: 4,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTasks(tt.input, tt.totalRooms)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseTasks() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUnfoldExcelData(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]string
		expected UnfoldedExcelData
		wantErr  bool
	}{
		{
			name: "正常系",
			input: [][]string{
				{"101, 103:105", "202, 204:207, 209:211"},
				{"フロア*4", "洗濯室*2"},
				{"", "自室清掃*?"},
			},
			expected: UnfoldedExcelData{
				roomNumbers: [][]int{
					{101, 103, 104, 105},
					{202, 204, 205, 206, 207, 209, 210, 211},
				},
				tasks: [][]string{
					{"フロア", "フロア", "フロア", "フロア"},
					{"洗濯室", "洗濯室", "自室清掃", "自室清掃", "自室清掃", "自室清掃", "自室清掃", "自室清掃"},
				},
			},
			wantErr: false,
		},
		{
			name:    "空のデータ",
			input:   [][]string{},
			wantErr: true,
		},
		{
			name: "エラー: パース失敗(部屋番号)",
			input: [][]string{
				{"101:100"},
				{"フロア*4"},
			},
			wantErr: true,
		},
		{
			name: "エラー: パース失敗(タスク)",
			input: [][]string{
				{"101, 102"},
				{"フロア*3"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unfoldExcelData(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unfoldExcelData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got.roomNumbers, tt.expected.roomNumbers) {
					t.Errorf("unfoldExcelData() roomNumbers = %v, want %v", got.roomNumbers, tt.expected.roomNumbers)
				}
				if !reflect.DeepEqual(got.tasks, tt.expected.tasks) {
					t.Errorf("unfoldExcelData() tasks = %v, want %v", got.tasks, tt.expected.tasks)
				}
			}
		})
	}
}

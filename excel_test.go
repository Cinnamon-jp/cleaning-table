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
			name: "空のデータ",
			input: [][]string{},
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

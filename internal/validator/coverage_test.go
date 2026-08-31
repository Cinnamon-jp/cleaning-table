package validator_test

import (
	"testing"

	"cleaning-table/internal/domain"
	"cleaning-table/internal/validator"
)

func TestCheckDuplicatesAndSparse(t *testing.T) {
	tests := []struct {
		name    string
		rules   []domain.AssignRule
		wantErr bool
	}{
		{
			name: "成功: すべての部屋が他のルールによって重複なくカバーされている",
			rules: []domain.AssignRule{
				{
					Rooms: []int{1, 2, 3}, // maxIdx (部屋数 3)
					Tasks: []string{"taskA", "taskB", "taskC"},
				},
				{
					Rooms: []int{1, 2},
					Tasks: []string{"taskA", "taskB"},
				},
				{
					Rooms: []int{3},
					Tasks: []string{"taskC"},
				},
			},
			wantErr: false,
		},
		{
			name: "エラー: 他のルール間で部屋に重複割り当てがある（部屋2が2回）",
			rules: []domain.AssignRule{
				{
					Rooms: []int{1, 2},
					Tasks: []string{"taskA", "taskB"},
				},
				{
					Rooms: []int{1, 2, 3, 4}, // maxIdx (部屋数 4)
					Tasks: []string{"taskA", "taskB", "taskC", "taskD"},
				},
				{
					Rooms: []int{2, 3, 4},
					Tasks: []string{"taskB", "taskC", "taskD"},
				},
			},
			wantErr: true, // 重複排除仕様によりエラー
		},
		{
			name: "エラー: カバーされていない部屋が残っている (部屋4が未カバー)",
			rules: []domain.AssignRule{
				{
					Rooms: []int{1, 2, 3, 4}, // maxIdx (部屋数 4)
					Tasks: []string{"taskA", "taskB", "taskC", "taskD"},
				},
				{
					Rooms: []int{1, 2},
					Tasks: []string{"taskA", "taskB"},
				},
				{
					Rooms: []int{3},
					Tasks: []string{"taskC"},
				},
			},
			wantErr: true,
		},
		{
			name: "エラー: ルールが全部屋ルール1つのみで引き算する対象がない",
			rules: []domain.AssignRule{
				{
					Rooms: []int{1, 2, 3},
					Tasks: []string{"taskA", "taskB", "taskC"},
				},
			},
			wantErr: true,
		},
		{
			name: "エラー: 1つの部分ルール内に重複部屋がある (101が2回)",
			rules: []domain.AssignRule{
				{
					Rooms: []int{101, 102, 103},
					Tasks: []string{"A", "B", "C"},
				},
				{
					Rooms: []int{101, 101},
					Tasks: []string{"A", "A"},
				},
				{
					Rooms: []int{102, 103},
					Tasks: []string{"B", "C"},
				},
			},
			wantErr: true,
		},
		{
			name: "エラー: 全部屋ルールに存在しない未知の部屋番号が含まれる (999)",
			rules: []domain.AssignRule{
				{
					Rooms: []int{101, 102},
					Tasks: []string{"A", "B"},
				},
				{
					Rooms: []int{101, 999},
					Tasks: []string{"A", "B"},
				},
				{
					Rooms: []int{102},
					Tasks: []string{"C"},
				},
			},
			wantErr: true,
		},
		{
			name: "エラー: 全部屋ルール自体に重複部屋番号が含まれる",
			rules: []domain.AssignRule{
				{
					Rooms: []int{101, 101, 102},
					Tasks: []string{"A", "B", "C"},
				},
				{
					Rooms: []int{101, 102},
					Tasks: []string{"A", "B"},
				},
			},
			wantErr: true,
		},
		{
			name:    "正常: 空のルールスライス",
			rules:   []domain.AssignRule{},
			wantErr: false,
		},
		{
			name:    "正常: rules が nil",
			rules:   nil,
			wantErr: false,
		},
		{
			name: "正常: すべてのルールの rooms が nil",
			rules: []domain.AssignRule{
				{Rooms: nil, Tasks: []string{"A"}},
				{Rooms: nil, Tasks: []string{"B"}},
			},
			wantErr: false,
		},
		{
			name: "正常: rooms が空スライスのみ",
			rules: []domain.AssignRule{
				{Rooms: []int{}, Tasks: []string{}},
			},
			wantErr: false,
		},
		{
			name: "エラー: 一部が nil で残りがカバーされていない",
			rules: []domain.AssignRule{
				{Rooms: nil, Tasks: []string{"A"}},
				{Rooms: []int{1, 2}, Tasks: []string{"A", "B"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.CheckDuplicatesAndSparse(tt.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDuplicatesAndSparse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

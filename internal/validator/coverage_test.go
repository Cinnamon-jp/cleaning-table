package validator_test

import (
	"testing"

	mapset "github.com/deckarep/golang-set/v2"

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
			name: "成功: すべての部屋が他のルールによってカバーされている",
			rules: []domain.AssignRule{
				{
					Rooms: mapset.NewSet(1, 2, 3), // maxIdx (部屋数 3)
					Tasks: []string{"taskA", "taskB", "taskC"},
				},
				{
					Rooms: mapset.NewSet(1, 2),
					Tasks: []string{"taskA", "taskB"},
				},
				{
					Rooms: mapset.NewSet(3),
					Tasks: []string{"taskC"},
				},
			},
			wantErr: false,
		},
		{
			name: "成功: 他のルール間で部屋に重複があってもカバーされていればOK",
			rules: []domain.AssignRule{
				{
					Rooms: mapset.NewSet(1, 2),
					Tasks: []string{"taskA", "taskB"},
				},
				{
					Rooms: mapset.NewSet(1, 2, 3, 4), // maxIdx (部屋数 4)
					Tasks: []string{"taskA", "taskB", "taskC", "taskD"},
				},
				{
					Rooms: mapset.NewSet(2, 3, 4),
					Tasks: []string{"taskB", "taskC", "taskD"},
				},
			},
			wantErr: false,
		},
		{
			name: "エラー: カバーされていない部屋が残っている",
			rules: []domain.AssignRule{
				{
					Rooms: mapset.NewSet(1, 2, 3, 4), // maxIdx (部屋数 4)
					Tasks: []string{"taskA", "taskB", "taskC", "taskD"},
				},
				{
					Rooms: mapset.NewSet(1, 2),
					Tasks: []string{"taskA", "taskB"},
				},
				{
					Rooms: mapset.NewSet(3),
					Tasks: []string{"taskC"},
				},
				// 部屋4がカバーされていない
			},
			wantErr: true,
		},
		{
			name: "エラー: ルールが全部屋ルール1つのみで引き算する対象がない",
			rules: []domain.AssignRule{
				{
					Rooms: mapset.NewSet(1, 2, 3),
					Tasks: []string{"taskA", "taskB", "taskC"},
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
			name: "正常: rooms が空集合（要素数0）のみ",
			rules: []domain.AssignRule{
				{Rooms: mapset.NewSet[int](), Tasks: []string{}},
			},
			wantErr: false,
		},
		{
			name: "エラー: 一部が nil で残りがカバーされていない",
			rules: []domain.AssignRule{
				{Rooms: nil, Tasks: []string{"A"}},
				{Rooms: mapset.NewSet(1, 2), Tasks: []string{"A", "B"}},
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

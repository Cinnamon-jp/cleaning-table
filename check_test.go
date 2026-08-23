package main

import (
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
)

func TestCheckDuplicatesAndSparse(t *testing.T) {
	tests := []struct {
		name    string
		rules   []assignRule
		wantErr bool
	}{
		{
			name: "成功: すべての部屋が他のルールによってカバーされている",
			rules: []assignRule{
				{
					rooms: mapset.NewSet(1, 2, 3), // maxIdx (部屋数 3)
					tasks: mapset.NewSet("taskA", "taskB", "taskC"),
				},
				{
					rooms: mapset.NewSet(1, 2),
					tasks: mapset.NewSet("taskA", "taskB"),
				},
				{
					rooms: mapset.NewSet(3),
					tasks: mapset.NewSet("taskC"),
				},
			},
			wantErr: false,
		},
		{
			name: "成功: 他のルール間で部屋に重複があってもカバーされていればOK",
			rules: []assignRule{
				{
					rooms: mapset.NewSet(1, 2),
					tasks: mapset.NewSet("taskA", "taskB"),
				},
				{
					rooms: mapset.NewSet(1, 2, 3, 4), // maxIdx (部屋数 4)
					tasks: mapset.NewSet("taskA", "taskB", "taskC", "taskD"),
				},
				{
					rooms: mapset.NewSet(2, 3, 4),
					tasks: mapset.NewSet("taskB", "taskC", "taskD"),
				},
			},
			wantErr: false,
		},
		{
			name: "エラー: カバーされていない部屋が残っている",
			rules: []assignRule{
				{
					rooms: mapset.NewSet(1, 2, 3, 4), // maxIdx (部屋数 4)
					tasks: mapset.NewSet("taskA", "taskB", "taskC", "taskD"),
				},
				{
					rooms: mapset.NewSet(1, 2),
					tasks: mapset.NewSet("taskA", "taskB"),
				},
				{
					rooms: mapset.NewSet(3),
					tasks: mapset.NewSet("taskC"),
				},
				// 部屋4がカバーされていない
			},
			wantErr: true,
		},
		{
			name: "エラー: ルールが全部屋ルール1つのみで引き算する対象がない",
			rules: []assignRule{
				{
					rooms: mapset.NewSet(1, 2, 3),
					tasks: mapset.NewSet("taskA", "taskB", "taskC"),
				},
			},
			wantErr: true,
		},
		{
			name:    "正常: 空のルールスライス",
			rules:   []assignRule{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkDuplicatesAndSparse(tt.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkDuplicatesAndSparse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

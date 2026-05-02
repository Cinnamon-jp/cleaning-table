package main

import (
	"math/rand/v2"
	"sort"
	"testing"
)

func TestShuffleAssign(t *testing.T) {
	data := &ExcelData{
		Columns: []ColumnData{
			{
				Rooms: []int{101, 102, 103, 104, 105},
				Roles: []Role{
					{Name: "ゴミ分別", Count: 2},
					{Name: "トイレ", Count: 3},
				},
			},
		},
	}

	// 乱数生成器のシードを固定して再現性を持たせる
	r := rand.New(rand.NewPCG(1, 2))

	result := ShuffleAssign(data, r)

	if len(result.Columns) != 1 {
		t.Fatalf("expected 1 column result, got %d", len(result.Columns))
	}

	colResult := result.Columns[0]
	if len(colResult.AssignedRoles) != 2 {
		t.Fatalf("expected 2 roles assigned, got %d", len(colResult.AssignedRoles))
	}

	var assignedRooms []int

	for _, ar := range colResult.AssignedRoles {
		switch ar.RoleName {
		case "ゴミ分別":
			if len(ar.Rooms) != 2 {
				t.Errorf("expected 2 rooms for ゴミ分別, got %d", len(ar.Rooms))
			}
		case "トイレ":
			if len(ar.Rooms) != 3 {
				t.Errorf("expected 3 rooms for トイレ, got %d", len(ar.Rooms))
			}
		}
		assignedRooms = append(assignedRooms, ar.Rooms...)
	}

	// 割り当てられた部屋の総数が一致するか、また重複がないかチェック
	if len(assignedRooms) != 5 {
		t.Fatalf("expected 5 rooms assigned in total, got %d", len(assignedRooms))
	}

	sort.Ints(assignedRooms)
	expectedRooms := []int{101, 102, 103, 104, 105}
	for i, r := range expectedRooms {
		if assignedRooms[i] != r {
			t.Errorf("expected room %d at %d, got %d", r, i, assignedRooms[i])
		}
	}
}

func TestShuffleAssign_NoRandomGenerator(t *testing.T) {
	data := &ExcelData{
		Columns: []ColumnData{
			{
				Rooms: []int{101, 102},
				Roles: []Role{
					{Name: "掃除", Count: 2},
				},
			},
		},
	}

	// r == nil でクラッシュしないことを確認する
	result := ShuffleAssign(data, nil)

	if len(result.Columns) != 1 {
		t.Fatalf("expected 1 column result")
	}
	if len(result.Columns[0].AssignedRoles[0].Rooms) != 2 {
		t.Fatalf("expected 2 rooms assigned")
	}
}

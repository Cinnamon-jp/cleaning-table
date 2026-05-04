package main

import (
	"testing"
)

func TestAssignTasks(t *testing.T) {
	data := UnfoldedExcelData{
		roomNumbers: [][]int{
			{101, 102, 103},
			{201, 202},
		},
		tasks: [][]string{
			{"TaskA", "TaskB", "TaskC"},
			{"TaskD", "TaskE"},
		},
	}

	result := assignTasks(data)

	// 列数の確認
	if len(result) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result))
	}

	// 各列の要素数の確認
	if len(result[0]) != 3 {
		t.Errorf("expected 3 assignments in col 0, got %d", len(result[0]))
	}
	if len(result[1]) != 2 {
		t.Errorf("expected 2 assignments in col 1, got %d", len(result[1]))
	}

	// 割り振られたタスクの集合が元のタスクの集合と一致するか確認する
	// column 0
	taskMap0 := make(map[string]int)
	for _, a := range result[0] {
		taskMap0[a.Task]++
	}
	for _, task := range data.tasks[0] {
		if taskMap0[task] == 0 {
			t.Errorf("task %s is missing or insufficient in col 0", task)
		} else {
			taskMap0[task]--
		}
	}

	// column 1
	taskMap1 := make(map[string]int)
	for _, a := range result[1] {
		taskMap1[a.Task]++
	}
	for _, task := range data.tasks[1] {
		if taskMap1[task] == 0 {
			t.Errorf("task %s is missing or insufficient in col 1", task)
		} else {
			taskMap1[task]--
		}
	}

	// 部屋番号がすべて含まれているか確認する
	// column 0
	roomMap0 := make(map[int]bool)
	for _, a := range result[0] {
		roomMap0[a.Room] = true
	}
	for _, room := range data.roomNumbers[0] {
		if !roomMap0[room] {
			t.Errorf("room %d is missing in col 0", room)
		}
	}

	// column 1
	roomMap1 := make(map[int]bool)
	for _, a := range result[1] {
		roomMap1[a.Room] = true
	}
	for _, room := range data.roomNumbers[1] {
		if !roomMap1[room] {
			t.Errorf("room %d is missing in col 1", room)
		}
	}
}

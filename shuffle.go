package main

import (
	"math/rand"
	"time"
)

// Assignment は部屋番号と割り当てられたタスクのペアを表します
type Assignment struct {
	Room int
	Task string
}

// assignTasks は展開されたエクセルデータを受け取り、各列ごとにタスクをシャッフルして部屋番号に割り当てます。
func assignTasks(data UnfoldedExcelData) [][]Assignment {
	var result [][]Assignment

	// 乱数生成器の初期化
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < len(data.roomNumbers); i++ {
		rooms := data.roomNumbers[i]
		tasks := data.tasks[i]
		
		// タスクスライスをコピーしてシャッフルする
		shuffledTasks := make([]string, len(tasks))
		copy(shuffledTasks, tasks)
		rng.Shuffle(len(shuffledTasks), func(i, j int) {
			shuffledTasks[i], shuffledTasks[j] = shuffledTasks[j], shuffledTasks[i]
		})

		var colAssignments []Assignment
		// 部屋数とタスク数が合わない場合に備えて安全にループする
		minLen := len(rooms)
		if len(shuffledTasks) < minLen {
			minLen = len(shuffledTasks)
		}

		for j := 0; j < minLen; j++ {
			colAssignments = append(colAssignments, Assignment{
				Room: rooms[j],
				Task: shuffledTasks[j],
			})
		}
		result = append(result, colAssignments)
	}

	return result
}

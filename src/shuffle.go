package src

import (
	"math/rand"
	"time"
)

// Assignment は部屋番号と割り当てられたタスクのペアを表します
type Assignment struct {
	Room int
	Task string
}

// AssignTasks は展開されたエクセルデータを受け取り、各列ごとにタスクをシャッフルして部屋番号に割り当てます。
func AssignTasks(data UnfoldedExcelData) [][]Assignment {
	var result [][]Assignment

	// 乱数生成器の初期化
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < len(data.RoomNumbers); i++ {
		rooms := data.RoomNumbers[i]
		tasks := data.Tasks[i]

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

// GroupByFloor は列ごとにグルーピングされた割り当て結果を、
// 階ごと（1F〜9F）にグルーピングし直した2次元スライスとして返します。
// 戻り値のインデックス 0 が 1F、インデックス 8 が 9F に対応します。
// 各階には01〜49号室すべてが号室番号昇順で含まれ、
// 割り当てがない部屋の Task は空文字列 "" となります。
func GroupByFloor(assignments [][]Assignment) [][]Assignment {
	const numFloors = 9
	const firstRoom = 1
	const lastRoom = 49

	// 各階01〜49号室の Assignment をあらかじめ生成する（Task は空文字列）
	result := make([][]Assignment, numFloors)
	for floor := 0; floor < numFloors; floor++ {
		floorNum := floor + 1
		floorAssignments := make([]Assignment, lastRoom-firstRoom+1)
		for room := firstRoom; room <= lastRoom; room++ {
			floorAssignments[room-firstRoom] = Assignment{
				Room: floorNum*100 + room,
				Task: "",
			}
		}
		result[floor] = floorAssignments
	}

	// 全列の全 Assignment を走査し、該当する階・号室の Task を上書きする
	for _, colAssignments := range assignments {
		for _, a := range colAssignments {
			floor := a.Room/100 - 1
			room := a.Room % 100

			if floor < 0 || floor >= numFloors {
				continue
			}
			if room < firstRoom || room > lastRoom {
				continue
			}

			result[floor][room-firstRoom].Task = a.Task
		}
	}

	return result
}

package main

import (
	"math/rand/v2"
)

// AssignedRole は役職と割り当てられた部屋のリストを保持します。
type AssignedRole struct {
	RoleName string
	Rooms    []int
}

// ColumnResult は1列分のシャッフルと割り当て結果を保持します。
type ColumnResult struct {
	AssignedRoles []AssignedRole
}

// ShuffleResult はファイル全体のシャッフル結果を保持します。
type ShuffleResult struct {
	Columns []ColumnResult
}

// ShuffleAssign は列ごとのデータを受け取り、ランダムに部屋をシャッフルして役職に割り当てます。
func ShuffleAssign(data *ExcelData, r *rand.Rand) *ShuffleResult {
	if r == nil {
		r = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}

	result := &ShuffleResult{
		Columns: make([]ColumnResult, 0, len(data.Columns)),
	}

	for _, col := range data.Columns {
		// 部屋のスライスをコピーしてシャッフル
		roomsCopy := make([]int, len(col.Rooms))
		copy(roomsCopy, col.Rooms)

		// シャッフル
		r.Shuffle(len(roomsCopy), func(i, j int) {
			roomsCopy[i], roomsCopy[j] = roomsCopy[j], roomsCopy[i]
		})

		var assignedRoles []AssignedRole
		currentIndex := 0

		for _, role := range col.Roles {
			count := role.Count
			// count分だけ切り出し
			assignedRooms := make([]int, count)
			copy(assignedRooms, roomsCopy[currentIndex:currentIndex+count])

			assignedRoles = append(assignedRoles, AssignedRole{
				RoleName: role.Name,
				Rooms:    assignedRooms,
			})
			currentIndex += count
		}

		result.Columns = append(result.Columns, ColumnResult{
			AssignedRoles: assignedRoles,
		})
	}

	return result
}

// Package matcher は過去の担当履歴をもとに均等なタスク割り当てを行います。
package matcher

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"cleaning-table/internal/algorithm"
	"cleaning-table/internal/domain"
)

// ShuffleTask は過去の担当履歴 (history) をもとに、
// 各部屋にタスクができるだけ均等に分配されるように割り当てを行います。
func ShuffleTask(
	rules []domain.AssignRule,
	history domain.AssignHistory,
) (
	result []domain.AssignResult,
	newHistory domain.AssignHistory,
	err error,
) {
	newHistory = history.Clone()
	result = make([]domain.AssignResult, 0)

	for i, rule := range rules {
		if rule.Rooms == nil || rule.Tasks == nil {
			return nil, nil, fmt.Errorf("rooms and tasks must not be nil (index %d)", i)
		}

		nRooms := rule.Rooms.Cardinality()
		nTasks := len(rule.Tasks)
		if nRooms != nTasks {
			return nil, nil, fmt.Errorf("number of rooms and tasks must match [%d != %d] (index %d)", nRooms, nTasks, i)
		}

		if nRooms == 0 {
			return nil, nil, fmt.Errorf("rooms and tasks must not be empty (index %d)", i)
		}

		rooms := rule.Rooms.ToSlice()
		tasks := slices.Clone(rule.Tasks)

		// 同一回数の候補間で偏りが出ないよう事前にランダムシャッフル
		//nolint:gosec // G404: 掃除当番の順序シャッフルであり暗号用途ではないため math/rand/v2 を使用
		rand.Shuffle(len(rooms), func(i, j int) {
			rooms[i], rooms[j] = rooms[j], rooms[i]
		})
		//nolint:gosec // G404: 掃除当番の順序シャッフルであり暗号用途ではないため math/rand/v2 を使用
		rand.Shuffle(len(tasks), func(i, j int) {
			tasks[i], tasks[j] = tasks[j], tasks[i]
		})

		// コスト行列を構築（担当回数の2乗をコストとする）
		cost := make([][]int, nRooms)
		for rIdx, room := range rooms {
			cost[rIdx] = make([]int, nTasks)
			for tIdx, task := range tasks {
				count := newHistory.GetCount(room, task)
				cost[rIdx][tIdx] = count * count
			}
		}

		// ハンガリアン法による最小重み完全マッチング
		matching := algorithm.MinWeightBipartiteMatching(cost)

		for rIdx, tIdx := range matching {
			room := rooms[rIdx]
			task := tasks[tIdx]

			result = append(result, domain.AssignResult{
				Room: room,
				Task: task,
			})

			newHistory.Increment(room, task)
		}
	}

	return result, newHistory, nil
}

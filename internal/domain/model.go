// Package domain は掃除当番割り当てのコアドメインモデルを提供します。
package domain

import (
	"maps"

	mapset "github.com/deckarep/golang-set/v2"
)

// AssignRule は割り当てルール（対象部屋集合とタスク一覧）を表します。
type AssignRule struct {
	Rooms mapset.Set[int]
	Tasks []string
}

// AssignResult は各部屋へのタスク割り当て結果を表します。
type AssignResult struct {
	Room int
	Task string
}

// AssignHistory は各部屋の過去のタスク担当回数（room -> task -> 回数）を保持します。
type AssignHistory map[int]map[string]int

// Clone は AssignHistory のディープコピーを作成します。
func (h AssignHistory) Clone() AssignHistory {
	if h == nil {
		return make(AssignHistory)
	}
	dst := make(AssignHistory, len(h))
	for room, tasks := range h {
		dst[room] = make(map[string]int, len(tasks))
		maps.Copy(dst[room], tasks)
	}
	return dst
}

// Increment は指定された部屋とタスクの担当回数を1増やします。
func (h AssignHistory) Increment(room int, task string) {
	if h[room] == nil {
		h[room] = make(map[string]int)
	}
	h[room][task]++
}

// GetCount は指定された部屋とタスクの担当回数を返します。
func (h AssignHistory) GetCount(room int, task string) int {
	if h == nil || h[room] == nil {
		return 0
	}
	return h[room][task]
}

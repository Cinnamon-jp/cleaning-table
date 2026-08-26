package main

import (
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"slices"
)

// shuffleTask は過去の担当履歴 (history) をもとに、
// 各部屋にタスクができるだけ均等に分配されるように割り当てを行う
func shuffleTask(
	rules []assignRule,
	history assignHistory,
) (
	result []assignResult,
	newHistory assignHistory,
	err error,
) {
	newHistory = copyHistory(history)
	result = make([]assignResult, 0)

	for i, rule := range rules {
		if rule.rooms == nil || rule.tasks == nil {
			return nil, nil, fmt.Errorf("rooms and tasks must not be nil (index %d)", i)
		}

		nRooms := rule.rooms.Cardinality()
		nTasks := len(rule.tasks)
		if nRooms != nTasks {
			return nil, nil, fmt.Errorf("number of rooms and tasks must match [%d != %d] (index %d)", nRooms, nTasks, i)
		}

		if nRooms == 0 {
			return nil, nil, fmt.Errorf("rooms and tasks must not be empty (index %d)", i)
		}

		rooms := rule.rooms.ToSlice()
		tasks := slices.Clone(rule.tasks)

		// 同一回数の候補間で偏りが出ないよう事前にランダムシャッフル
		rand.Shuffle(len(rooms), func(i, j int) {
			rooms[i], rooms[j] = rooms[j], rooms[i]
		})
		rand.Shuffle(len(tasks), func(i, j int) {
			tasks[i], tasks[j] = tasks[j], tasks[i]
		})

		// コスト行列を構築
		cost := make([][]int, nRooms)
		for rIdx, room := range rooms {
			cost[rIdx] = make([]int, nTasks)
			for tIdx, task := range tasks {
				count := 0
				if taskCounts, ok := newHistory[room]; ok {
					count = taskCounts[task]
				}
				cost[rIdx][tIdx] = count * count // コストを2乗で定義する
			}
		}

		// ハンガリアン法による最小重み完全マッチング
		matching := minWeightBipartiteMatching(cost)

		for rIdx, tIdx := range matching {
			room := rooms[rIdx]
			task := tasks[tIdx]

			result = append(result, assignResult{
				room: room,
				task: task,
			})

			if newHistory[room] == nil {
				newHistory[room] = make(map[string]int)
			}
			newHistory[room][task]++
		}
	}

	return result, newHistory, nil
}

// copyHistory は assignHistory のディープコピーを作成する
func copyHistory(src assignHistory) assignHistory {
	dst := make(assignHistory)
	for room, tasks := range src {
		dst[room] = make(map[string]int, len(tasks))
		maps.Copy(dst[room], tasks)
	}
	return dst
}

// minWeightBipartiteMatching はハンガリアン法を用いて、
// N x N コスト行列に対する最小コストのマッチングを求める
func minWeightBipartiteMatching(cost [][]int) []int {
	n := len(cost)
	if n == 0 {
		return nil
	}

	u := make([]int, n+1)
	v := make([]int, n+1)
	p := make([]int, n+1)
	way := make([]int, n+1)

	const inf = math.MaxInt / 2

	for i := 1; i <= n; i++ {
		p[0] = i
		j0 := 0
		minv := make([]int, n+1)
		for k := range minv {
			minv[k] = inf
		}
		used := make([]bool, n+1)

		for {
			used[j0] = true
			i0 := p[j0]
			delta := inf
			j1 := 0

			for j := 1; j <= n; j++ {
				if !used[j] {
					cur := cost[i0-1][j-1] - u[i0] - v[j]
					if cur < minv[j] {
						minv[j] = cur
						way[j] = j0
					}
					if minv[j] < delta {
						delta = minv[j]
						j1 = j
					}
				}
			}

			for j := 0; j <= n; j++ {
				if used[j] {
					u[p[j]] += delta
					v[j] -= delta
				} else {
					minv[j] -= delta
				}
			}

			j0 = j1
			if p[j0] == 0 {
				break
			}
		}

		for {
			j1 := way[j0]
			p[j0] = p[j1]
			j0 = j1
			if j0 == 0 {
				break
			}
		}
	}

	result := make([]int, n)
	for j := 1; j <= n; j++ {
		if p[j] > 0 {
			result[p[j]-1] = j - 1
		}
	}
	return result
}

// Package algorithm はマッチング等の汎用アルゴリズムを提供します。
package algorithm

import (
	"math"
)

// MinWeightBipartiteMatching はハンガリアン法を用いて、
// N x N コスト行列に対する最小コストの完全マッチングを求めます。
// 行 i に対して割り当てられた列インデックス matching[i] のスライスを返します。
func MinWeightBipartiteMatching(cost [][]int) []int {
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

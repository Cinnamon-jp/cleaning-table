package algorithm_test

import (
	"reflect"
	"testing"

	"cleaning-table/internal/algorithm"
)

func TestMinWeightBipartiteMatching(t *testing.T) {
	tests := []struct {
		name     string
		cost     [][]int
		expected []int
	}{
		{
			name:     "空行列",
			cost:     [][]int{},
			expected: nil,
		},
		{
			name: "1x1 行列",
			cost: [][]int{
				{5},
			},
			expected: []int{0},
		},
		{
			name: "2x2 行列 (対角線が最小)",
			cost: [][]int{
				{1, 10},
				{10, 2},
			},
			expected: []int{0, 1},
		},
		{
			name: "2x2 行列 (反対角線が最小)",
			cost: [][]int{
				{10, 1},
				{2, 10},
			},
			expected: []int{1, 0},
		},
		{
			name: "3x3 行列",
			cost: [][]int{
				{4, 1, 3},
				{2, 0, 5},
				{3, 2, 2},
			},
			expected: []int{1, 0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := algorithm.MinWeightBipartiteMatching(tt.cost)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("MinWeightBipartiteMatching() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMinWeightBipartiteMatching_UniformCost(t *testing.T) {
	// 全要素が同じコスト（0や100など）の場合でも、
	// すべての行と列が1対1で完全マッチング（重複なし）されること
	sizes := []int{1, 4, 10}
	for _, n := range sizes {
		for _, uniformVal := range []int{0, 100} {
			cost := make([][]int, n)
			for i := range n {
				cost[i] = make([]int, n)
				for j := range n {
					cost[i][j] = uniformVal
				}
			}

			matching := algorithm.MinWeightBipartiteMatching(cost)
			if len(matching) != n {
				t.Fatalf("expected matching length %d, got %d", n, len(matching))
			}

			usedCols := make(map[int]bool)
			for row, col := range matching {
				if col < 0 || col >= n {
					t.Errorf("row %d assigned invalid col %d", row, col)
				}
				if usedCols[col] {
					t.Errorf("col %d was assigned multiple times", col)
				}
				usedCols[col] = true
			}
		}
	}
}

func TestMinWeightBipartiteMatching_LargeScale(t *testing.T) {
	// 50x50 および 100x100 の大規模行列に対する完全マッチングの検証
	sizes := []int{50, 100}

	for _, n := range sizes {
		cost := make([][]int, n)
		for i := range n {
			cost[i] = make([]int, n)
			for j := range n {
				cost[i][j] = (i + j) % 10
			}
		}

		matching := algorithm.MinWeightBipartiteMatching(cost)
		if len(matching) != n {
			t.Fatalf("size %d: expected matching length %d, got %d", n, n, len(matching))
		}

		usedCols := make(map[int]bool, n)
		for _, col := range matching {
			if usedCols[col] {
				t.Fatalf("size %d: duplicate column %d assigned", n, col)
			}
			usedCols[col] = true
		}
	}
}

func BenchmarkMinWeightBipartiteMatching_50x50(b *testing.B) {
	const n = 50
	cost := make([][]int, n)
	for i := range n {
		cost[i] = make([]int, n)
		for j := range n {
			cost[i][j] = (i*7 + j*13) % 100
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_ = algorithm.MinWeightBipartiteMatching(cost)
	}
}

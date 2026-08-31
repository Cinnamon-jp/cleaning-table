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
			// Row 0 -> Col 1 (1)
			// Row 1 -> Col 0 (2)
			// Row 2 -> Col 2 (2)
			// Total cost = 1 + 2 + 2 = 5
			// matching: [2, 0, 1] (Row 0 -> Col 2(3), Row 1 -> Col 1(0), Row 2 -> Col 0(3) = 6) vs (Row 0->Col 2(3)+Row 1->Col 0(2)+Row 2->Col 1(2) = 7)
			// Row 0->Col 2 (3), Row 1->Col 1 (0), Row 2->Col 0 (3) -> 6
			// Row 0->Col 1 (1) [not possible because row 1 col 1 is 0], wait:
			// Row 0->1(1), Row 1->0(2), Row 2->2(2) -> total 1+2+2 = 5 -> matching is [1, 0, 2]
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

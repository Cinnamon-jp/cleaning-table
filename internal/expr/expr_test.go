package expr_test

import (
	"fmt"
	"reflect"
	"testing"

	"cleaning-table/internal/expr"
)

func helperLookup(env map[string][]int) expr.LookupFunc {
	return func(ident string) (expr.Set, error) {
		items, ok := env[ident]
		if !ok {
			return nil, fmt.Errorf("undefined identifier: %s", ident)
		}
		return expr.NewSet(items...), nil
	}
}

func TestEval_BasicOperations(t *testing.T) {
	env := map[string][]int{
		"all_rooms.1F": {101, 102, 103, 104, 105},
		"facility":     {101, 109},
		"posts":        {103},
		"A":            {1, 2, 3, 4},
		"B":            {3, 4, 5, 6},
		"C":            {4, 6, 7},
	}
	lookup := helperLookup(env)

	tests := []struct {
		name     string
		expr     string
		expected []int
	}{
		{
			name:     "単一識別子（ドット付き）",
			expr:     "all_rooms.1F",
			expected: []int{101, 102, 103, 104, 105},
		},
		{
			name:     "和集合 (A + B)",
			expr:     "A + B",
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "差集合 (A - B)",
			expr:     "A - B",
			expected: []int{1, 2},
		},
		{
			name:     "積集合 (A & B)",
			expr:     "A & B",
			expected: []int{3, 4},
		},
		{
			name:     "連続差集合 (A - B - C)",
			expr:     "A - B - C",
			expected: []int{1, 2},
		},
		{
			name:     "ideal_test.yaml 実例パターン (all_rooms.1F - facility - posts)",
			expr:     "all_rooms.1F - facility - posts",
			expected: []int{102, 104, 105},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expr.Eval(tt.expr, lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Eval(%q) = %v, want %v", tt.expr, got, tt.expected)
			}
		})
	}
}

func TestEval_PrecedenceAndParentheses(t *testing.T) {
	env := map[string][]int{
		"A": {1, 2, 3, 4},
		"B": {3, 4, 5},
		"C": {4, 5, 6},
		"D": {1, 5, 7},
	}
	lookup := helperLookup(env)

	tests := []struct {
		name     string
		expr     string
		expected []int
	}{
		{
			name:     "積集合が和集合より優先 (A + B & C = A + (B & C))",
			expr:     "A + B & C",
			expected: []int{1, 2, 3, 4, 5}, // B & C = {4, 5}. A + {4, 5} = {1, 2, 3, 4, 5}
		},
		{
			name:     "括弧による優先度変更 ((A + B) & C)",
			expr:     "(A + B) & C",
			expected: []int{4, 5}, // A + B = {1, 2, 3, 4, 5}. {1, 2, 3, 4, 5} & {4, 5, 6} = {4, 5}
		},
		{
			name:     "積集合が差集合より優先 (A - B & C = A - (B & C))",
			expr:     "A - B & C",
			expected: []int{1, 2, 3}, // B & C = {4, 5}. A \ {4, 5} = {1, 2, 3}
		},
		{
			name:     "括弧による差集合優先 ((A - B) & C)",
			expr:     "(A - B) & C",
			expected: []int{}, // A \ B = {1, 2}. {1, 2} & {4, 5, 6} = {}
		},
		{
			name:     "括弧の入れ子 ((A + (B & C)) - D)",
			expr:     "(A + (B & C)) - D",
			expected: []int{2, 3, 4}, // B & C = {4, 5}. A + {4, 5} = {1, 2, 3, 4, 5}. \ {1, 5, 7} = {2, 3, 4}
		},
		{
			name:     "複合式 ((A + B) - (C & D))",
			expr:     "(A + B) - (C & D)",
			expected: []int{1, 2, 3, 4}, // A + B = {1, 2, 3, 4, 5}. C & D = {5}. \ {5} = {1, 2, 3, 4}
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expr.Eval(tt.expr, lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Eval(%q) = %v, want %v", tt.expr, got, tt.expected)
			}
		})
	}
}

func TestEval_Errors(t *testing.T) {
	env := map[string][]int{
		"A": {1, 2},
		"B": {2, 3},
	}
	lookup := helperLookup(env)

	tests := []struct {
		name string
		expr string
	}{
		{name: "空文字列", expr: ""},
		{name: "空白文字のみ", expr: "   \t\n  "},
		{name: "末尾が演算子", expr: "A + "},
		{name: "先頭が演算子", expr: "+ A"},
		{name: "連続する演算子", expr: "A + + B"},
		{name: "演算子なしの連続識別子", expr: "A B"},
		{name: "閉じられていない括弧", expr: "(A + B"},
		{name: "余分な閉じ括弧", expr: "A + B)"},
		{name: "空の括弧", expr: "()"},
		{name: "未定義のシンボル", expr: "A + unknown"},
		{name: "不正な文字", expr: "A @ B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := expr.Eval(tt.expr, lookup)
			if err == nil {
				t.Errorf("Eval(%q) expected error, got nil", tt.expr)
			}
		})
	}

	t.Run("lookup が nil の場合", func(t *testing.T) {
		_, err := expr.Eval("A + B", nil)
		if err == nil {
			t.Error("expected error when lookup is nil, got nil")
		}
	})
}

func TestSetMethods(t *testing.T) {
	s1 := expr.NewSet(1, 2, 3)
	clone := s1.Clone()
	if !reflect.DeepEqual(s1.ToSlice(), clone.ToSlice()) {
		t.Errorf("clone mismatch: got %v, want %v", clone.ToSlice(), s1.ToSlice())
	}

	var nilSet expr.Set
	clonedNil := nilSet.Clone()
	if clonedNil == nil || len(clonedNil) != 0 {
		t.Errorf("expected empty non-nil set from nil Clone, got %v", clonedNil)
	}
}

func BenchmarkEval_Complex(b *testing.B) {
	env := map[string][]int{
		"all_rooms.1F": makeRange(101, 150),
		"all_rooms.2F": makeRange(201, 250),
		"facility":     {101, 102, 109, 202, 241},
		"posts.mentor": {103, 201, 211, 227},
		"posts.safety": {110, 125, 203, 213},
	}
	lookup := helperLookup(env)
	expression := "(all_rooms.1F + all_rooms.2F) - facility - (posts.mentor + posts.safety)"

	b.ResetTimer()
	for b.Loop() {
		_, err := expr.Eval(expression, lookup)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func makeRange(start, end int) []int {
	a := make([]int, end-start+1)
	for i := range a {
		a[i] = start + i
	}
	return a
}

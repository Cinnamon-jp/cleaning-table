package domain_test

import (
	"testing"

	"cleaning-table/internal/domain"
)

func TestAssignHistory_Clone(t *testing.T) {
	orig := domain.AssignHistory{
		101: {"A": 1, "B": 2},
		102: {"C": 3},
	}

	clone := orig.Clone()

	// 複製されていること
	if clone.GetCount(101, "A") != 1 || clone.GetCount(101, "B") != 2 || clone.GetCount(102, "C") != 3 {
		t.Fatalf("unexpected cloned content: %+v", clone)
	}

	// 変更が元のデータに波及しないこと
	clone.Increment(101, "A")
	if orig.GetCount(101, "A") != 1 {
		t.Errorf("original was mutated: got %d, want 1", orig.GetCount(101, "A"))
	}
	if clone.GetCount(101, "A") != 2 {
		t.Errorf("clone was not incremented: got %d, want 2", clone.GetCount(101, "A"))
	}
}

func TestAssignHistory_NilHandling(t *testing.T) {
	var h domain.AssignHistory

	// nil に対する Clone
	cloned := h.Clone()
	if cloned == nil {
		t.Error("expected non-nil map from nil Clone")
	}

	// nil に対する GetCount
	if count := h.GetCount(101, "A"); count != 0 {
		t.Errorf("expected 0 from nil GetCount, got %d", count)
	}

	// 初期化して Increment
	h = make(domain.AssignHistory)
	h.Increment(101, "TaskX")
	if count := h.GetCount(101, "TaskX"); count != 1 {
		t.Errorf("expected 1 after increment, got %d", count)
	}
}

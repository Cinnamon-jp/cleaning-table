package storage_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"cleaning-table/internal/domain"
	"cleaning-table/internal/storage"
)

func TestLoadHistory_NotExist(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "non_existent.json")

	h, err := storage.LoadHistory(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil || len(h) != 0 {
		t.Errorf("expected empty non-nil history, got %v", h)
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "sub", "history.json")

	orig := domain.AssignHistory{
		101: {"フロア": 2, "ゴミ分別": 1},
		202: {"自室清掃": 5},
	}

	if err := storage.SaveHistory(orig, path); err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	loaded, err := storage.LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	if !reflect.DeepEqual(orig, loaded) {
		t.Errorf("loaded history mismatch: got %+v, want %+v", loaded, orig)
	}
}

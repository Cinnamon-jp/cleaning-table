package output_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cleaning-table/internal/domain"
	"cleaning-table/internal/output"
)

func TestGenerateReverseText(t *testing.T) {
	results := []domain.AssignResult{
		{Room: 104, Task: "フロア"},
		{Room: 204, Task: "フロア"},
		{Room: 105, Task: "トイレ1期"},
		{Room: 106, Task: "自室清掃"},
		{Room: 205, Task: "フロア"},
	}

	text := output.GenerateReverseText(results)

	// 各タスクの見出しが含まれていること
	if !strings.Contains(text, "【フロア】 (計 3 部屋)") {
		t.Errorf("missing floor task header in: %s", text)
	}
	if !strings.Contains(text, "【トイレ1期】 (計 1 部屋)") {
		t.Errorf("missing toilet task header in: %s", text)
	}
	if !strings.Contains(text, "【自室清掃】 (計 1 部屋)") {
		t.Errorf("missing private room task header in: %s", text)
	}

	// フロア別に整形されていること (2F: 204, 205)
	if !strings.Contains(text, "2F: 204, 205") {
		t.Errorf("missing floor 2 grouping in: %s", text)
	}
	if !strings.Contains(text, "1F: 104") {
		t.Errorf("missing floor 1 grouping in: %s", text)
	}
}

func TestWriteReverseText(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "sub", "reverse.txt")

	results := []domain.AssignResult{
		{Room: 101, Task: "TaskA"},
	}

	if err := output.WriteReverseText(results, outputPath); err != nil {
		t.Fatalf("WriteReverseText failed: %v", err)
	}

	//nolint:gosec // G304: テスト出力ファイル読み込みのため
	data, err := os.ReadFile(filepath.Clean(outputPath))
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if !strings.Contains(string(data), "【TaskA】 (計 1 部屋)") {
		t.Errorf("unexpected content: %s", string(data))
	}
}

func TestWritePDF(t *testing.T) {
	fontPath := "../../NotoSerifJP-Bold.ttf"
	if _, err := os.Stat(fontPath); err != nil {
		fontPath = "../../NotoSerifJP-VariableFont_wght.ttf"
		if _, err := os.Stat(fontPath); err != nil {
			t.Skipf("skipping PDF test, font not found: %v", err)
		}
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_table.pdf")

	results := []domain.AssignResult{
		{Room: 101, Task: "寮長"},
		{Room: 102, Task: "指導寮生"},
		{Room: 103, Task: "フロア"},
		{Room: 104, Task: "トイレ1期"},
		{Room: 105, Task: "自室清掃"},
		{Room: 201, Task: "シャワー1"},
		{Room: 202, Task: "自室清掃"},
	}

	if err := output.WritePDF(results, fontPath, outputPath); err != nil {
		t.Fatalf("WritePDF failed: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("failed to stat generated PDF: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("generated PDF is empty (0 bytes)")
	}
}

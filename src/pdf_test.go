package src

import (
	"os"
	"strings"
	"testing"
)

func TestFormatRoomNumber(t *testing.T) {
	tests := []struct {
		name     string
		room     int
		expected string
	}{
		{
			name:     "1F01号室",
			room:     101,
			expected: "01",
		},
		{
			name:     "2F49号室",
			room:     249,
			expected: "49",
		},
		{
			name:     "9F01号室",
			room:     901,
			expected: "01",
		},
		{
			name:     "5F25号室",
			room:     525,
			expected: "25",
		},
		{
			name:     "1F10号室",
			room:     110,
			expected: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRoomNumber(tt.room)
			if result != tt.expected {
				t.Errorf("formatRoomNumber(%d) = %q, expected %q", tt.room, result, tt.expected)
			}
		})
	}
}

func TestBuildFloorPage(t *testing.T) {
	t.Run("1F: ページが正しく構築される", func(t *testing.T) {
		// テスト用の49部屋分のAssignmentを作成
		assignments := make([]Assignment, 49)
		for i := 0; i < 49; i++ {
			assignments[i] = Assignment{
				Room: 100 + i + 1,
				Task: "テストタスク",
			}
		}

		p := buildFloorPage(1, assignments)

		if p == nil {
			t.Fatal("buildFloorPage returned nil")
		}

		rows := p.GetRows()
		if rows == nil {
			t.Fatal("GetRows returned nil")
		}

		// 期待される行数:
		// タイトル行(1) + 間隔行(1) + ヘッダー行(1) + データ行(30) = 33行
		expectedRows := 33
		if len(rows) != expectedRows {
			t.Errorf("expected %d rows, got %d", expectedRows, len(rows))
		}
	})

	t.Run("9F: 最後の階でも正しく構築される", func(t *testing.T) {
		assignments := make([]Assignment, 49)
		for i := 0; i < 49; i++ {
			assignments[i] = Assignment{
				Room: 900 + i + 1,
				Task: "タスク",
			}
		}

		p := buildFloorPage(9, assignments)

		if p == nil {
			t.Fatal("buildFloorPage returned nil")
		}

		rows := p.GetRows()
		expectedRows := 33
		if len(rows) != expectedRows {
			t.Errorf("expected %d rows, got %d", expectedRows, len(rows))
		}
	})

	t.Run("タスクが空文字列の部屋がある場合", func(t *testing.T) {
		assignments := make([]Assignment, 49)
		for i := 0; i < 49; i++ {
			task := ""
			if i%3 == 0 {
				task = "フロア"
			}
			assignments[i] = Assignment{
				Room: 300 + i + 1,
				Task: task,
			}
		}

		p := buildFloorPage(3, assignments)

		if p == nil {
			t.Fatal("buildFloorPage returned nil")
		}

		rows := p.GetRows()
		expectedRows := 33
		if len(rows) != expectedRows {
			t.Errorf("expected %d rows, got %d", expectedRows, len(rows))
		}
	})
}

func TestGeneratePDF(t *testing.T) {
	t.Run("フォントファイルが存在しない場合はエラー", func(t *testing.T) {
		// テスト用のfloorAssignments（9階分）
		floorAssignments := makeTestFloorAssignments()

		err := GeneratePDF(floorAssignments, "nonexistent_font.ttf")
		if err == nil {
			t.Error("expected error for nonexistent font file, got nil")
		}
	})

	t.Run("正常系: PDFファイルが生成される", func(t *testing.T) {
		// ipaexg.ttf がプロジェクトルートに存在することを前提とする
		fontPath := "ipaexg.ttf"
		if _, err := os.Stat(fontPath); os.IsNotExist(err) {
			// src/配下で実行された場合を想定して親ディレクトリも探索
			fontPath = "../ipaexg.ttf"
			if _, err := os.Stat(fontPath); os.IsNotExist(err) {
				t.Skipf("font file %s not found, skipping test", fontPath)
			}
		}

		floorAssignments := makeTestFloorAssignments()

		err := GeneratePDF(floorAssignments, fontPath)
		if err != nil {
			t.Fatalf("generatePDF returned error: %v", err)
		}

		// 生成されたPDFファイルを検索して確認
		entries, err := os.ReadDir(".")
		if err != nil {
			t.Fatalf("failed to read directory: %v", err)
		}

		var generatedFile string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "output_") && strings.HasSuffix(entry.Name(), ".pdf") {
				generatedFile = entry.Name()
			}
		}

		if generatedFile == "" {
			t.Fatal("no PDF file was generated")
		}

		// ファイルサイズが0でないことを確認
		info, err := os.Stat(generatedFile)
		if err != nil {
			t.Fatalf("failed to stat generated file: %v", err)
		}
		if info.Size() == 0 {
			t.Error("generated PDF file is empty")
		}

		// テスト後のクリーンアップ
		t.Cleanup(func() {
			if err := os.Remove(generatedFile); err != nil {
				t.Logf("failed to remove generated file: %v", err)
			}
		})
	})
}

// makeTestFloorAssignments はテスト用の9階分のfloorAssignmentsを生成します。
func makeTestFloorAssignments() [][]Assignment {
	tasks := []string{"フロア", "トイレ", "ゴミ分別", "洗濯室", "階段"}

	floorAssignments := make([][]Assignment, 9)
	for floor := 0; floor < 9; floor++ {
		floorNum := floor + 1
		floorAssignments[floor] = make([]Assignment, 49)
		for room := 1; room <= 49; room++ {
			floorAssignments[floor][room-1] = Assignment{
				Room: floorNum*100 + room,
				Task: tasks[room%len(tasks)],
			}
		}
	}

	return floorAssignments
}

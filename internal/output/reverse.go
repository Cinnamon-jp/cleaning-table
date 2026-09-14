// Package output は割り当て結果の逆引きテキストおよび PDF 当番表の出力を提供します。
package output

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"cleaning-table/internal/domain"
)

// GenerateReverseText は AssignResult スライスからタスク別の逆引きテキストを生成します。
func GenerateReverseText(results []domain.AssignResult) string {
	// タスク -> 部屋スライスのマッピング
	taskRooms := make(map[string][]int)
	// タスクの出現順（またはソート順）を保持
	taskOrder := make([]string, 0)
	seenTask := make(map[string]bool)

	for _, res := range results {
		if !seenTask[res.Task] {
			seenTask[res.Task] = true
			taskOrder = append(taskOrder, res.Task)
		}
		taskRooms[res.Task] = append(taskRooms[res.Task], res.Room)
	}

	// タスク名を五十音／アルファベット順にソート（安定した出力のため）
	slices.Sort(taskOrder)

	var sb strings.Builder
	sb.WriteString("======================================================================\n")
	sb.WriteString("  掃除当番表 逆引きリスト（タスク別担当部屋一覧）\n")
	sb.WriteString("======================================================================\n\n")

	for _, task := range taskOrder {
		rooms := taskRooms[task]
		slices.Sort(rooms)

		fmt.Fprintf(&sb, "【%s】 (計 %d 部屋)\n", task, len(rooms))

		// フロアごとにグループ化 (room / 100)
		floorMap := make(map[int][]int)
		floorOrder := make([]int, 0)
		seenFloor := make(map[int]bool)

		for _, room := range rooms {
			fl := room / 100
			if !seenFloor[fl] {
				seenFloor[fl] = true
				floorOrder = append(floorOrder, fl)
			}
			floorMap[fl] = append(floorMap[fl], room)
		}
		slices.Sort(floorOrder)

		for _, fl := range floorOrder {
			flRooms := floorMap[fl]
			roomStrs := make([]string, len(flRooms))
			for i, r := range flRooms {
				roomStrs[i] = strconv.Itoa(r)
			}
			fmt.Fprintf(&sb, "  %dF: %s\n", fl, strings.Join(roomStrs, ", "))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// WriteReverseText はタスク別逆引きテキストを指定されたファイルに出力します。
func WriteReverseText(results []domain.AssignResult, outputPath string) error {
	cleanPath := filepath.Clean(outputPath)
	dir := filepath.Dir(cleanPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	content := GenerateReverseText(results)
	//nolint:gosec // G306: 出力テキストファイルは通常の権限で書き込み
	if err := os.WriteFile(cleanPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write reverse text to %q: %w", outputPath, err)
	}

	return nil
}

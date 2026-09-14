package output

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/signintech/gopdf"

	"cleaning-table/internal/domain"
)

// WritePDF は AssignResult スライスをもとに掲示用の PDF 当番表を生成します。
func WritePDF(results []domain.AssignResult, fontPath, outputPath string) error {
	cleanFontPath := filepath.Clean(fontPath)
	if _, err := os.Stat(cleanFontPath); err != nil {
		return fmt.Errorf("font file not found at %q: %w", fontPath, err)
	}

	cleanOutputPath := filepath.Clean(outputPath)
	dir := filepath.Dir(cleanOutputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	// 部屋番号順にソートし、フロアごとに分類
	sortedResults := slices.Clone(results)
	slices.SortFunc(sortedResults, func(a, b domain.AssignResult) int {
		return a.Room - b.Room
	})

	floorResults := make(map[int][]domain.AssignResult)
	floorKeys := make([]int, 0)
	seenFloor := make(map[int]bool)

	for _, res := range sortedResults {
		fl := res.Room / 100
		if !seenFloor[fl] {
			seenFloor[fl] = true
			floorKeys = append(floorKeys, fl)
		}
		floorResults[fl] = append(floorResults[fl], res)
	}
	slices.Sort(floorKeys)

	pdf := gopdf.GoPdf{}
	// A4 横向き (841.89 x 595.28 pt)
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4Landscape})

	const fontName = "custom_font"
	if err := pdf.AddTTFFont(fontName, cleanFontPath); err != nil {
		return fmt.Errorf("failed to load TTF font %q: %w", fontPath, err)
	}

	const (
		marginL       = 35.0
		marginT       = 25.0
		pageW         = 841.89
		usableW       = pageW - (marginL * 2) // 771.89
		floorsPerPage = 3
		cols          = 5
		rowH          = 13.0
		colW          = usableW / float64(cols) // 154.38
		roomW         = 38.0
		taskW         = colW - roomW // 116.38
	)

	// 1ページに最大 3 フロアずつ配置
	for i := 0; i < len(floorKeys); i += floorsPerPage {
		pdf.AddPage()

		endIdx := min(i+floorsPerPage, len(floorKeys))
		pageFloors := floorKeys[i:endIdx]

		currentY := marginT

		// ヘッダー描画
		if err := pdf.SetFont(fontName, "", 15); err != nil {
			return err
		}
		pdf.SetTextColor(30, 41, 59)
		pdf.SetXY(marginL, currentY)
		if err := pdf.Text("掃除当番表"); err != nil {
			return err
		}

		// 日時とページ番号
		if err := pdf.SetFont(fontName, "", 9); err != nil {
			return err
		}
		pdf.SetTextColor(100, 116, 139)
		dateStr := fmt.Sprintf("出力日時: %s | ページ %d / %d",
			time.Now().Format("2006-01-02 15:04"), (i/floorsPerPage)+1, (len(floorKeys)+floorsPerPage-1)/floorsPerPage)
		pdf.SetXY(pageW-marginL-250, currentY+4)
		if err := pdf.CellWithOption(&gopdf.Rect{W: 250, H: 15}, dateStr, gopdf.CellOption{Align: gopdf.Right}); err != nil {
			return err
		}

		currentY += 22

		// 各フロアの描画
		for _, fl := range pageFloors {
			items := floorResults[fl]

			// フロア見出し
			if err := pdf.SetFont(fontName, "", 11); err != nil {
				return err
			}
			pdf.SetTextColor(15, 23, 42)
			pdf.SetXY(marginL, currentY)
			if err := pdf.Text(fmt.Sprintf("■ %d階 (%dF)  [全 %d 部屋]", fl, fl, len(items))); err != nil {
				return err
			}

			currentY += 16

			// グリッド表描画 (5列×行数)
			numRows := (len(items) + cols - 1) / cols
			for row := range numRows {
				for col := range cols {
					idx := row*cols + col
					if idx >= len(items) {
						continue
					}
					item := items[idx]

					cellX := marginL + float64(col)*colW
					cellY := currentY + float64(row)*rowH

					// 背景と枠線
					pdf.SetLineWidth(0.4)
					if item.Task == "自室清掃" {
						pdf.SetFillColor(248, 250, 252) // 非常に薄いグレー
						pdf.SetStrokeColor(226, 232, 240)
					} else {
						pdf.SetFillColor(241, 245, 249) // やや強調
						pdf.SetStrokeColor(203, 213, 225)
					}
					if err := pdf.Rectangle(cellX, cellY, cellX+colW, cellY+rowH, "DF", 0, 0); err != nil {
						return err
					}

					// 部屋番号セル
					if err := pdf.SetFont(fontName, "", 8.5); err != nil {
						return err
					}
					pdf.SetTextColor(71, 85, 105)
					pdf.SetXY(cellX+2, cellY+1.5)
					if err := pdf.CellWithOption(&gopdf.Rect{W: roomW - 2, H: rowH}, fmt.Sprintf("%d", item.Room), gopdf.CellOption{Align: gopdf.Left}); err != nil {
						return err
					}

					// タスク名セル
					if item.Task == "自室清掃" {
						pdf.SetTextColor(100, 116, 139)
					} else {
						pdf.SetTextColor(15, 23, 42) // 濃い色
					}
					pdf.SetXY(cellX+roomW, cellY+1.5)
					if err := pdf.CellWithOption(&gopdf.Rect{W: taskW - 2, H: rowH}, item.Task, gopdf.CellOption{Align: gopdf.Left}); err != nil {
						return err
					}
				}
			}

			currentY += float64(numRows)*rowH + 18
		}
	}

	if err := pdf.WritePdf(cleanOutputPath); err != nil {
		return fmt.Errorf("failed to write PDF to %q: %w", outputPath, err)
	}

	return nil
}

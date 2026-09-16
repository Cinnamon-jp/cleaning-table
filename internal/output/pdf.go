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

	// 各部屋のタスクをマップ化し、フロア一覧と部屋番号の範囲（オフセット）を算出
	taskMap := make(map[int]string, len(results))
	floorSet := make(map[int]bool)
	minOffset := 1
	maxOffset := 1

	for _, res := range results {
		taskMap[res.Room] = res.Task
		fl := res.Room / 100
		floorSet[fl] = true
		off := res.Room % 100
		if off > maxOffset {
			maxOffset = off
		}
	}

	floorKeys := make([]int, 0, len(floorSet))
	for fl := range floorSet {
		floorKeys = append(floorKeys, fl)
	}
	slices.Sort(floorKeys)

	// 全フロアを同じ部屋番号範囲（固定長）で揃え、役職のない部屋はタスクを空文字にする
	floorResults := make(map[int][]domain.AssignResult, len(floorKeys))
	for _, fl := range floorKeys {
		items := make([]domain.AssignResult, 0, maxOffset-minOffset+1)
		for off := minOffset; off <= maxOffset; off++ {
			room := fl*100 + off
			items = append(items, domain.AssignResult{
				Room: room,
				Task: taskMap[room],
			})
		}
		floorResults[fl] = items
	}

	pdf := gopdf.GoPdf{}
	// A4 縦向き (595.28 x 841.89 pt)
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	const fontName = "custom_font"
	if err := pdf.AddTTFFont(fontName, cleanFontPath); err != nil {
		return fmt.Errorf("failed to load TTF font %q: %w", fontPath, err)
	}

	const (
		marginL       = 35.0
		marginT       = 30.0
		pageW         = 595.28                 // A4 縦向きの幅
		usableW       = pageW - (marginL * 2)  // 525.28
		floorsPerPage = 1                      // 1ページに 1 フロア配置
		colGap        = 25.0                   // 2列間の間隔
		colW          = (usableW - colGap) / 2 // 250.14
		roomW         = 46.0                   // 部屋番号セルの幅
		taskW         = colW - roomW           // 204.14 (タスク名セルの幅)
		rowH          = 22.0                   // 1行の高さ（30行で660pt）
		splitRoomOff  = 30                     // 左列: 1〜30号室、右列: 31号室〜
	)

	// 1ページに 1 フロアずつ配置
	for i := 0; i < len(floorKeys); i += floorsPerPage {
		pdf.AddPage()

		endIdx := min(i+floorsPerPage, len(floorKeys))
		pageFloors := floorKeys[i:endIdx]

		currentY := marginT

		// ヘッダー描画
		if err := pdf.SetFont(fontName, "", 16); err != nil {
			return err
		}
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(marginL, currentY)
		if err := pdf.Text("清掃割り振り表"); err != nil {
			return err
		}

		// 日時とページ番号
		if err := pdf.SetFont(fontName, "", 9); err != nil {
			return err
		}
		pdf.SetTextColor(0, 0, 0)
		dateStr := fmt.Sprintf("出力日時: %s",
			time.Now().Format("2006-01-02 15:04"))
		pdf.SetXY(pageW-marginL-250, currentY+4)
		if err := pdf.CellWithOption(&gopdf.Rect{W: 250, H: 15}, dateStr, gopdf.CellOption{Align: gopdf.Right}); err != nil {
			return err
		}

		currentY += 26

		// 各フロアの描画
		for _, fl := range pageFloors {
			items := floorResults[fl]

			// フロア見出し
			if err := pdf.SetFont(fontName, "", 18); err != nil {
				return err
			}
			pdf.SetTextColor(0, 0, 0)
			pdf.SetXY(marginL, currentY)
			if err := pdf.CellWithOption(&gopdf.Rect{W: usableW, H: 26}, fmt.Sprintf("%d階", fl), gopdf.CellOption{Align: gopdf.Center | gopdf.Middle}); err != nil {
				return err
			}

			currentY += 50

			// 2列レイアウト描画（左列: 01〜30号室、右列: 31〜49号室）
			for _, item := range items {
				off := item.Room % 100
				var col, row int
				if off <= splitRoomOff {
					col = 0
					row = off - 1
				} else {
					col = 1
					row = off - (splitRoomOff + 1)
				}

				cellX := marginL + float64(col)*(colW+colGap)
				cellY := currentY + float64(row)*rowH

				// 背景と枠線（全セル白背景・黒枠線）
				pdf.SetLineWidth(0.5)
				pdf.SetFillColor(255, 255, 255)
				pdf.SetStrokeColor(0, 0, 0)
				if err := pdf.Rectangle(cellX, cellY, cellX+colW, cellY+rowH, "DF", 0, 0); err != nil {
					return err
				}

				// 部屋番号とタスク名の仕切り線
				pdf.Line(cellX+roomW, cellY, cellX+roomW, cellY+rowH)

				// 部屋番号セル
				if err := pdf.SetFont(fontName, "", 10); err != nil {
					return err
				}
				pdf.SetTextColor(0, 0, 0)
				pdf.SetXY(cellX, cellY)
				if err := pdf.CellWithOption(&gopdf.Rect{W: roomW, H: rowH}, fmt.Sprintf("%d", item.Room), gopdf.CellOption{Align: gopdf.Center | gopdf.Middle}); err != nil {
					return err
				}

				// タスク名セル
				if err := pdf.SetFont(fontName, "", 10); err != nil {
					return err
				}
				pdf.SetTextColor(0, 0, 0)
				pdf.SetXY(cellX+roomW+6, cellY)
				if err := pdf.CellWithOption(&gopdf.Rect{W: taskW - 8, H: rowH}, item.Task, gopdf.CellOption{Align: gopdf.Left | gopdf.Middle}); err != nil {
					return err
				}
			}

			currentY += float64(splitRoomOff)*rowH + 18
		}
	}

	if err := pdf.WritePdf(cleanOutputPath); err != nil {
		return fmt.Errorf("failed to write PDF to %q: %w", outputPath, err)
	}

	return nil
}

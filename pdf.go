package main

import (
	"cleaning-table/util"
	"fmt"
	"os"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/page"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// generatePDF はシャッフル結果を受け取り、掃除当番表のPDFファイルを生成します。
// 1ページごとに1階分を記載し、合計9ページ作成します。
func generatePDF(floorAssignments [][]Assignment, fontPath string) error {
	// フォントファイルの読み込み
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		util.Logger.Error(
			"pdf.go: os.ReadFile()",
			fmt.Sprintf("Failed to read font file: %v", err),
			fmt.Sprintf("フォントファイルの読み込みに失敗しました: %v", err),
		)
		return fmt.Errorf("failed to read font file %s: %w", fontPath, err)
	}

	// カスタムフォントの定義（Normal と Bold の両方に同じTTFファイルを使用）
	customFonts := []*entity.CustomFont{
		{
			Family: "ipaexg",
			Style:  fontstyle.Normal,
			Bytes:  fontBytes,
		},
		{
			Family: "ipaexg",
			Style:  fontstyle.Bold,
			Bytes:  fontBytes,
		},
	}

	// Maroto設定の構築
	cfg := config.NewBuilder().
		WithCustomFonts(customFonts).
		WithDefaultFont(&props.Font{
			Family: "ipaexg",
			Size:   10,
			Style:  fontstyle.Normal,
		}).
		Build()

	// Marotoインスタンスの作成
	m := maroto.New(cfg)

	// 9ページ分のページを生成（1F〜9F）
	for floor := 0; floor < len(floorAssignments); floor++ {
		floorPage := buildFloorPage(floor+1, floorAssignments[floor])
		m.AddPages(floorPage)
	}

	// PDFの生成
	doc, err := m.Generate()
	if err != nil {
		util.Logger.Error(
			"pdf.go: maroto.Generate()",
			fmt.Sprintf("Failed to generate PDF: %v", err),
			fmt.Sprintf("PDFの生成に失敗しました: %v", err),
		)
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// ファイル名の生成（タイムスタンプ形式）
	fileName := fmt.Sprintf("output_%s.pdf", time.Now().Format("20060102150405"))

	// ファイルとして保存
	err = doc.Save(fileName)
	if err != nil {
		util.Logger.Error(
			"pdf.go: doc.Save()",
			fmt.Sprintf("Failed to save PDF file: %v", err),
			fmt.Sprintf("PDFファイルの保存に失敗しました: %v", err),
		)
		return fmt.Errorf("failed to save PDF file %s: %w", fileName, err)
	}

	util.Logger.Info(
		"pdf.go: generatePDF()",
		fmt.Sprintf("PDF file saved successfully: %s", fileName),
		fmt.Sprintf("PDFファイルを保存しました: %s", fileName),
	)

	return nil
}

// buildFloorPage は1階分のページ内容を構築します。
// floor は階数（1〜9）、assignments はその階の全49部屋分の Assignment スライスです。
func buildFloorPage(floor int, assignments []Assignment) core.Page {
	// ボーダー付きセルのスタイル
	borderedCellStyle := &props.Cell{
		BorderType:      border.Full,
		BorderThickness: 0.2,
		BorderColor:     &props.Color{Red: 0, Green: 0, Blue: 0},
	}

	// ヘッダーテキストのスタイル（太字、中央揃え）
	headerTextProps := props.Text{
		Size:  9,
		Style: fontstyle.Bold,
		Align: align.Center,
		Top:   1,
	}

	// データセルのテキストスタイル（中央揃え）
	dataCellTextProps := props.Text{
		Size:  9,
		Align: align.Center,
		Top:   1,
	}

	// タイトル行
	titleRow := row.New(15).Add(
		col.New(12).Add(
			text.New(fmt.Sprintf("%dF掃除当番表", floor), props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Align: align.Center,
				Top:   2,
			}),
		),
	)

	// 間隔行
	spacerRow := row.New(5)

	// ページ作成
	p := page.New()
	p.Add(titleRow)
	p.Add(spacerRow)

	// ヘッダー行（左テーブル・右テーブル共通）
	headerRow := row.New(7).Add(
		col.New(2).WithStyle(borderedCellStyle).Add(
			text.New("部屋番号", headerTextProps),
		),
		col.New(3).WithStyle(borderedCellStyle).Add(
			text.New("掃除場所", headerTextProps),
		),
		col.New(2), // 間隔列（ボーダーなし）
		col.New(2).WithStyle(borderedCellStyle).Add(
			text.New("部屋番号", headerTextProps),
		),
		col.New(3).WithStyle(borderedCellStyle).Add(
			text.New("掃除場所", headerTextProps),
		),
	)
	p.Add(headerRow)

	// データ行の生成
	// 左テーブル: 01〜30号室（インデックス 0〜29）= 30行
	// 右テーブル: 31〜49号室（インデックス 30〜48）= 19行
	// 最大行数は30行。右テーブルは19行目以降を空欄とする。
	const leftTableRows = 30
	const rightTableStartIndex = 30
	const rightTableRows = 19

	for i := 0; i < leftTableRows; i++ {
		// 左テーブルのデータ
		leftRoomStr := formatRoomNumber(assignments[i].Room)
		leftTask := assignments[i].Task

		// 右テーブルのデータ（範囲外の場合は空欄）
		rightRoomStr := ""
		rightTask := ""
		rightIndex := rightTableStartIndex + i
		hasRightData := rightIndex < len(assignments) && i < rightTableRows

		if hasRightData {
			rightRoomStr = formatRoomNumber(assignments[rightIndex].Room)
			rightTask = assignments[rightIndex].Task
		}

		// 右テーブルのセルスタイル（データがある場合はボーダー付き、ない場合はボーダーなし）
		var rightRoomStyle *props.Cell
		var rightTaskStyle *props.Cell
		if hasRightData {
			rightRoomStyle = borderedCellStyle
			rightTaskStyle = borderedCellStyle
		}

		dataRow := row.New(7).Add(
			// 左テーブル: 部屋番号
			col.New(2).WithStyle(borderedCellStyle).Add(
				text.New(leftRoomStr, dataCellTextProps),
			),
			// 左テーブル: 掃除場所
			col.New(3).WithStyle(borderedCellStyle).Add(
				text.New(leftTask, dataCellTextProps),
			),
			// 間隔列
			col.New(2),
			// 右テーブル: 部屋番号
			col.New(2).WithStyle(rightRoomStyle).Add(
				text.New(rightRoomStr, dataCellTextProps),
			),
			// 右テーブル: 掃除場所
			col.New(3).WithStyle(rightTaskStyle).Add(
				text.New(rightTask, dataCellTextProps),
			),
		)
		p.Add(dataRow)
	}

	return p
}

// formatRoomNumber は部屋番号の下2桁を2桁ゼロ埋めでフォーマットします。
// 例: 101 → "01", 249 → "49"
func formatRoomNumber(room int) string {
	return fmt.Sprintf("%02d", room%100)
}

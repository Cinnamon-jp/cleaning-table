// Package pdf はシャッフルされた掃除当番データをPDFファイルに出力する処理を格納する
package pdf

import (
	"cleaning-table/src/model"
	"cleaning-table/src/util"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/signintech/gopdf"
)

// レイアウト定数
// マージン
const (
	marginLeft = 30.0
	marginTop  = 40.0
)

// タイトル
const (
	titleFontSize     = 20.0
	titleMarginBottom = 25.0
)

// テーブル
const (
	tableFontSize   = 11.0
	headerFontSize  = 12.0
	cellHeight      = 20.0
	roomNumColWidth = 80.0
	postColWidth    = 150.0
	tableGap        = 25.0
	cellPaddingLeft = 8.0
	cellPaddingTop  = 5.0
	headerBgGray    = 217 // RGB値 (0-255)
)

// OutputPdf はシャッフルされた掃除当番データをPDFファイルに出力する。
// 1ページに1階分のデータを記載し、データが存在する階のみページを作成する。
// 各ページには「XF掃除当番表」のタイトルと、01〜29号室・30〜49号室の2テーブルを横並びで配置する。
func OutputPdf(data []model.ShuffledPostSet) error {
	// 階数別にデータをグルーピング
	floorMap := groupByFloor(data)

	// 階数の昇順でソートされたキーリストを作成
	floors := sortedFloorKeys(floorMap)

	if len(floors) == 0 {
		util.Logger(util.Warn, "output_pdf.go/OutputPdf()", "No data to output / データが空のためPDFを生成しません", "データが空のためPDFを生成しません")
		return nil
	}

	// PDF初期化
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// フォント読み込み
	if err := loadFont(&pdf); err != nil {
		util.Logger(util.Error, "output_pdf.go/OutputPdf()/loadFont()", "Error when loading font", "フォントの読み込み中にエラーが発生しました")
		return err
	}

	// 各階のページを生成
	for _, floor := range floors {
		entries := floorMap[floor]
		// 部屋番号でソート
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].RoomNumber < entries[j].RoomNumber
		})

		if err := renderFloorPage(&pdf, floor, entries); err != nil {
			util.Logger(util.Error, "output_pdf.go/OutputPdf()/renderFloorPage()", fmt.Sprintf("Error when rendering page for %dF", floor), fmt.Sprintf("%dFのページ生成中にエラーが発生しました", floor))
			return err
		}
	}

	// PDF保存
	outputFileName := util.Input("保存するPDFファイル名を入力してください: ") + ".pdf"
	if err := pdf.WritePdf(outputFileName); err != nil {
		util.Logger(util.Error, "output_pdf.go/OutputPdf()/pdf.WritePdf()", "Error when writing PDF file", "PDFファイルの書き込み中にエラーが発生しました")
		return err
	}

	util.Logger(util.Info, "output_pdf.go/OutputPdf()", fmt.Sprintf("PDF file created: %s", outputFileName), fmt.Sprintf("PDFファイルを作成しました: %s", outputFileName))
	return nil
}

// groupByFloor はShuffledPostSetのスライスを階数ごとにグルーピングする。
// 部屋番号の百の位を階数として扱う（例: 101 → 1F, 302 → 3F）。
func groupByFloor(data []model.ShuffledPostSet) map[int][]model.ShuffledPostSet {
	floorMap := make(map[int][]model.ShuffledPostSet)
	for _, entry := range data {
		floor := entry.RoomNumber / 100
		floorMap[floor] = append(floorMap[floor], entry)
	}
	return floorMap
}

// sortedFloorKeys はfloorMapのキーを昇順ソートして返す。
func sortedFloorKeys(floorMap map[int][]model.ShuffledPostSet) []int {
	floors := make([]int, 0, len(floorMap))
	for floor := range floorMap {
		floors = append(floors, floor)
	}
	sort.Ints(floors)
	return floors
}

// loadFont はカレントディレクトリ内のTTFフォントファイルを探索し、PDFオブジェクトに読み込む。
// TTFファイルが複数見つかった場合は util.Select を使用してユーザーに選択させる。
// TTFファイルが1つだけの場合はそのまま使用する。
func loadFont(pdf *gopdf.GoPdf) error {
	// カレントディレクトリ内の.ttfファイルを探索
	ttfFiles, err := filepath.Glob("*.ttf")
	if err != nil {
		util.Logger(util.Error, "output_pdf.go/loadFont()/filepath.Glob()", "Error when searching for TTF files", "TTFファイルの探索中にエラーが発生しました")
		return err
	}

	if len(ttfFiles) == 0 {
		return fmt.Errorf("no TTF font files found in current directory / カレントディレクトリにTTFフォントファイルが見つかりません")
	}

	// フォントファイルの選択
	var selectedFont string
	if len(ttfFiles) > 1 {
		if selectedFont, err = util.Select("使用するフォントファイルを選択してください", ttfFiles); err != nil {
			util.Logger(util.Error, "output_pdf.go/loadFont()/util.Select()", "Error when selecting font file", "フォントファイルの選択中にエラーが発生しました")
			return err
		}
	} else {
		selectedFont = ttfFiles[0]
	}

	// フォントの読み込み
	if err := pdf.AddTTFFont("japanese", selectedFont); err != nil {
		util.Logger(util.Error, "output_pdf.go/loadFont()/pdf.AddTTFFont()", fmt.Sprintf("Error when loading font: %s", selectedFont), fmt.Sprintf("フォントの読み込みに失敗しました: %s", selectedFont))
		return err
	}

	util.Logger(util.Info, "output_pdf.go/loadFont()", fmt.Sprintf("Font loaded: %s", selectedFont), fmt.Sprintf("フォントを読み込みました: %s", selectedFont))
	return nil
}

// renderFloorPage は1階分のページを描画する。
func renderFloorPage(pdf *gopdf.GoPdf, floor int, entries []model.ShuffledPostSet) error {
	pdf.AddPage()

	// タイトル描画
	if err := drawTitle(pdf, floor); err != nil {
		return err
	}

	// 01〜29号室と30〜49号室にデータを分割
	leftEntries, rightEntries := splitEntries(floor, entries)

	// テーブル開始Y座標
	tableStartY := marginTop + titleFontSize + titleMarginBottom

	// 左テーブル描画（01〜29号室）
	leftTableX := marginLeft
	if err := drawTable(pdf, leftTableX, tableStartY, leftEntries); err != nil {
		return err
	}

	// 右テーブル描画（30〜49号室）
	rightTableX := marginLeft + roomNumColWidth + postColWidth + tableGap
	if err := drawTable(pdf, rightTableX, tableStartY, rightEntries); err != nil {
		return err
	}

	return nil
}

// drawTitle はページ上部中央にタイトルを描画する。
func drawTitle(pdf *gopdf.GoPdf, floor int) error {
	if err := pdf.SetFont("japanese", "", titleFontSize); err != nil {
		return err
	}

	title := fmt.Sprintf("%dF掃除当番表", floor)

	// タイトルのテキスト幅を計測して中央揃え
	titleWidth, err := pdf.MeasureTextWidth(title)
	if err != nil {
		return err
	}

	pageWidth := gopdf.PageSizeA4.W
	titleX := (pageWidth - titleWidth) / 2

	pdf.SetXY(titleX, marginTop)
	if err := pdf.Cell(nil, title); err != nil {
		return err
	}

	return nil
}

// splitEntries は部屋番号の下2桁に基づいて、01〜29号室と30〜49号室にデータを分割する。
func splitEntries(floor int, entries []model.ShuffledPostSet) ([]model.ShuffledPostSet, []model.ShuffledPostSet) {
	var left, right []model.ShuffledPostSet
	for _, entry := range entries {
		roomSuffix := entry.RoomNumber - floor*100
		if roomSuffix >= 1 && roomSuffix <= 29 {
			left = append(left, entry)
		} else if roomSuffix >= 30 && roomSuffix <= 49 {
			right = append(right, entry)
		}
	}
	return left, right
}

// drawTable は指定座標にテーブルを描画する。
// ヘッダー行（「部屋番号」「掃除場所」）と、それに続くデータ行を描画する。
func drawTable(pdf *gopdf.GoPdf, startX, startY float64, entries []model.ShuffledPostSet) error {
	currentY := startY

	// ヘッダー行の描画
	if err := drawHeaderRow(pdf, startX, currentY); err != nil {
		return err
	}
	currentY += cellHeight

	// データ行の描画
	if err := pdf.SetFont("japanese", "", tableFontSize); err != nil {
		return err
	}

	for _, entry := range entries {
		roomStr := fmt.Sprintf("%02d", entry.RoomNumber%100)

		// 部屋番号セル
		if err := drawCell(pdf, startX, currentY, roomNumColWidth, cellHeight, roomStr); err != nil {
			return err
		}

		// 掃除場所セル
		if err := drawCell(pdf, startX+roomNumColWidth, currentY, postColWidth, cellHeight, entry.Post); err != nil {
			return err
		}

		currentY += cellHeight
	}

	return nil
}

// drawHeaderRow はテーブルのヘッダー行を描画する。
// グレーの背景色付きでヘッダーテキスト（「部屋番号」「掃除場所」）を描画する。
func drawHeaderRow(pdf *gopdf.GoPdf, startX, startY float64) error {
	if err := pdf.SetFont("japanese", "", headerFontSize); err != nil {
		return err
	}

	// ヘッダー背景色（グレー）を描画
	pdf.SetFillColor(headerBgGray, headerBgGray, headerBgGray)

	// 部屋番号ヘッダーの背景
	if err := pdf.Rectangle(startX, startY, startX+roomNumColWidth, startY+cellHeight, "F", 0, 0); err != nil {
		return err
	}
	// 掃除場所ヘッダーの背景
	if err := pdf.Rectangle(startX+roomNumColWidth, startY, startX+roomNumColWidth+postColWidth, startY+cellHeight, "F", 0, 0); err != nil {
		return err
	}

	// テキスト色を黒にリセット
	pdf.SetTextColor(0, 0, 0)

	// 部屋番号ヘッダーテキスト
	if err := drawCell(pdf, startX, startY, roomNumColWidth, cellHeight, "部屋番号"); err != nil {
		return err
	}

	// 掃除場所ヘッダーテキスト
	if err := drawCell(pdf, startX+roomNumColWidth, startY, postColWidth, cellHeight, "掃除場所"); err != nil {
		return err
	}

	// 塗りつぶし色をリセット（白）
	pdf.SetFillColor(255, 255, 255)

	return nil
}

// drawCell は指定座標にセル（矩形の罫線とテキスト）を描画する。
func drawCell(pdf *gopdf.GoPdf, x, y, width, height float64, text string) error {
	// セルの罫線を描画
	pdf.SetLineWidth(0.5)
	pdf.SetStrokeColor(0, 0, 0)
	if err := pdf.Rectangle(x, y, x+width, y+height, "D", 0, 0); err != nil {
		return err
	}

	// テキストを描画（パディングあり）
	pdf.SetXY(x+cellPaddingLeft, y+cellPaddingTop)
	if err := pdf.Cell(nil, text); err != nil {
		return err
	}

	return nil
}

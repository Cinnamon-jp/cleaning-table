package main

import (
	"log"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func main() {
	// 1. 設定の構築（デフォルトでA4サイズが適用されます）
	cfg := config.NewBuilder().Build()

	// 2. Marotoインスタンスの作成
	m := maroto.New(cfg)

	// 3. 行(Row)と列(Col)を追加してコンポーネント(Text)を配置
	m.AddRows(
		// 高さ20mmの行を作成
		row.New(20).Add(
			// 幅12（つまり横幅いっぱい）の列を作成
			col.New(12).Add(
				// テキストを追加し、文字サイズとスタイル（太字）を指定
				text.New("Hello, Maroto v2!", props.Text{
					Size:  20,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// 4. PDFの生成
	doc, err := m.Generate()
	if err != nil {
		log.Fatal("PDF生成エラー:", err)
	}

	// 5. ファイルとして保存
	err = doc.Save("basic.pdf")
	if err != nil {
		log.Fatal("保存エラー:", err)
	}
}

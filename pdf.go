package main

import (
	"fmt"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// GeneratePDF はシャッフル結果を受け取り、指定されたファイルパスにPDFを生成します。
func GeneratePDF(result *ShuffleResult, outputPath string) error {
	builder := config.NewBuilder().
		WithCustomFonts([]*entity.CustomFont{
			{
				Family: "NotoSansJP",
				Style:  fontstyle.Normal,
				Bytes:  NotoSansJPRegularBytes,
			},
			{
				Family: "NotoSansJP",
				Style:  fontstyle.Bold,
				Bytes:  NotoSansJPRegularBytes, // ボールドも同じフォントを使う
			},
		}).
		WithDefaultFont(&props.Font{
			Family: "NotoSansJP",
			Style:  fontstyle.Normal,
			Size:   12,
		})

	m := maroto.New(builder.Build())

	// タイトル
	m.AddRows(
		row.New(20).Add(
			text.NewCol(12, "掃除当番表", props.Text{
				Top:   5,
				Style: fontstyle.Bold,
				Size:  20,
				Align: align.Center,
			}),
		),
		row.New(10), // 空白行
	)

	// 各列のデータを描画
	for i, colResult := range result.Columns {
		// 列（グループ）の見出し
		m.AddRows(
			row.New(10).Add(
				text.NewCol(12, fmt.Sprintf("グループ %d", i+1), props.Text{
					Style: fontstyle.Bold,
					Size:  14,
				}),
			),
		)

		for _, ar := range colResult.AssignedRoles {
			roomStrs := make([]string, len(ar.Rooms))
			for j, r := range ar.Rooms {
				roomStrs[j] = fmt.Sprintf("%d", r)
			}
			roomsText := strings.Join(roomStrs, ", ")

			m.AddRows(
				row.New(8).Add(
					text.NewCol(4, ar.RoleName, props.Text{
						Size: 12,
					}),
					text.NewCol(8, roomsText, props.Text{
						Size: 12,
					}),
				),
			)
		}

		m.AddRows(row.New(5)) // グループ間の空白
	}

	doc, err := m.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	err = doc.Save(outputPath)
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}

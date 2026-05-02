package main

import (
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// mapRolesToRooms は ShuffleResult から 部屋番号 -> 役職名 のマップを作成します。
func mapRolesToRooms(result *ShuffleResult) map[int]string {
	roleMap := make(map[int]string)
	for _, col := range result.Columns {
		for _, ar := range col.AssignedRoles {
			for _, room := range ar.Rooms {
				roleMap[room] = ar.RoleName
			}
		}
	}
	return roleMap
}

// GeneratePDF はシャッフル結果を受け取り、指定されたファイルパスにPDFを生成します。
func GeneratePDF(result *ShuffleResult, outputPath string) error {
	builder := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithOrientation(orientation.Horizontal).
		WithMaxGridSize(11). // 両端に部屋番号(2枠) + 1F~9F(9枠) = 11枠
		WithLeftMargin(10).
		WithRightMargin(10).
		WithTopMargin(5).
		WithBottomMargin(5).
		WithCustomFonts([]*entity.CustomFont{
			{
				Family: "NotoSansJP",
				Style:  fontstyle.Normal,
				Bytes:  NotoSansJPRegularBytes,
			},
			{
				Family: "NotoSansJP",
				Style:  fontstyle.Bold,
				Bytes:  NotoSansJPRegularBytes,
			},
		}).
		WithDefaultFont(&props.Font{
			Family: "NotoSansJP",
			Style:  fontstyle.Normal,
			Size:   6, // 1ページに収めるためにフォントサイズを小さくする
		})

	m := maroto.New(builder.Build())

	// タイトル
	m.AddRows(
		row.New(8).Add(
			text.NewCol(11, "掃除当番表", props.Text{
				Top:   1,
				Style: fontstyle.Bold,
				Size:  12,
				Align: align.Center,
			}),
		),
	)

	roomToRole := mapRolesToRooms(result)

	cellStyle := &props.Cell{
		BorderType:      border.Full,
		BorderThickness: 0.1,
	}

	// ヘッダー行 (空白, 1F, 2F, ..., 9F, 空白)
	headerCols := []core.Col{
		text.NewCol(1, "", props.Text{
			Align: align.Center,
		}).WithStyle(cellStyle),
	}
	for f := 1; f <= 9; f++ {
		headerCols = append(headerCols, text.NewCol(1, fmt.Sprintf("%dF", f), props.Text{
			Align: align.Center,
			Style: fontstyle.Bold,
			Top:   1,
		}).WithStyle(cellStyle))
	}
	headerCols = append(headerCols, text.NewCol(1, "", props.Text{
		Align: align.Center,
	}).WithStyle(cellStyle))

	m.AddRows(row.New(4).Add(headerCols...))

	// 各行 (01 ~ 49)
	for r := 1; r <= 49; r++ {
		rowCols := []core.Col{
			text.NewCol(1, fmt.Sprintf("%02d", r), props.Text{
				Align: align.Center,
				Top:   1,
			}).WithStyle(cellStyle),
		}

		for f := 1; f <= 9; f++ {
			roomNum := f*100 + r
			roleName := roomToRole[roomNum]
			if roleName == "" {
				roleName = "-"
			}
			rowCols = append(rowCols, text.NewCol(1, roleName, props.Text{
				Align: align.Center,
				Top:   1,
			}).WithStyle(cellStyle))
		}

		rowCols = append(rowCols, text.NewCol(1, fmt.Sprintf("%02d", r), props.Text{
			Align: align.Center,
			Top:   1,
		}).WithStyle(cellStyle))

		m.AddRows(row.New(3.5).Add(rowCols...))
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

package main

import (
	"os"
	"testing"
)

func TestGeneratePDF(t *testing.T) {
	// テスト用のモックデータ（複数階、多数の部屋を含む）
	mockResult := &ShuffleResult{
		Columns: []ColumnResult{
			{
				AssignedRoles: []AssignedRole{
					{RoleName: "ゴミ分別", Rooms: []int{101, 202, 303, 404}},
					{RoleName: "トイレ", Rooms: []int{105, 110, 510, 949}},
				},
			},
			{
				AssignedRoles: []AssignedRole{
					{RoleName: "フロア", Rooms: []int{201, 302, 403, 848}},
				},
			},
		},
	}

	outputPath := "test_output.pdf"

	// PDF生成
	err := GeneratePDF(mockResult, outputPath)
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to be created", outputPath)
	}

	// クリーンアップ
	if err := os.Remove(outputPath); err != nil {
		t.Logf("Failed to clean up file %s: %v", outputPath, err)
	}
}

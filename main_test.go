package main

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestFindExcelFiles(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir := t.TempDir()

	// 現在の作業ディレクトリを保存
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在の作業ディレクトリの取得に失敗しました: %v", err)
	}
	// テスト終了後に元のディレクトリに戻るようにする
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("元の作業ディレクトリへの復元に失敗しました: %v", err)
		}
	}()

	// 作業ディレクトリを一時ディレクトリに変更
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("作業ディレクトリの変更に失敗しました: %v", err)
	}

	// テスト用のファイルを作成
	filesToCreate := []string{
		"test1.xlsx",
		"test2.xlsx",
		"document.txt",
		"image.png",
	}

	for _, fileName := range filesToCreate {
		file, err := os.Create(fileName)
		if err != nil {
			t.Fatalf("テストファイル %s の作成に失敗しました: %v", fileName, err)
		}
		file.Close()
	}

	// ディレクトリであるが、拡張子が .xlsx のものを作成 (無視されることを確認するため)
	if err := os.Mkdir("dummy_dir.xlsx", 0755); err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	// 関数を実行
	files, err := findExcelFiles()
	if err != nil {
		t.Fatalf("findExcelFiles がエラーを返しました: %v", err)
	}

	// 期待される結果
	expected := []string{"test1.xlsx", "test2.xlsx"}

	// 順序に依存しないようにソートする
	sort.Strings(files)
	sort.Strings(expected)

	// 結果を検証
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("期待されるファイルリスト %v に対して、実際は %v でした", expected, files)
	}
}

package util

import (
	"bytes"
	"io"
	"testing"
)

// mockReadCloser は標準入力をモックするための構造体です。
type mockReadCloser struct {
	io.Reader
}

func (m mockReadCloser) Close() error { return nil }

// mockWriteCloser は標準出力をモックするための構造体です。
type mockWriteCloser struct {
	io.Writer
}

func (m mockWriteCloser) Close() error { return nil }

func TestChooseOne(t *testing.T) {
	tests := []struct {
		name        string
		list        []string
		input       string // シミュレートするキーボード入力
		wantResult  string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "空のリストの場合はエラー",
			list:        []string{},
			input:       "",
			wantResult:  "",
			wantErr:     true,
			expectedErr: "the list is empty",
		},
		{
			name:       "最初の要素を選択",
			list:       []string{"Apple", "Banana", "Cherry"},
			input:      "\r", // Enter キー
			wantResult: "Apple",
			wantErr:    false,
		},
		{
			name:       "二番目の要素を選択",
			list:       []string{"Apple", "Banana", "Cherry"},
			input:      "\x1b[B\r", // 下矢印キー + Enter キー
			wantResult: "Banana",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != "" {
				// 標準入出力をモック
				chooseStdin = mockReadCloser{Reader: bytes.NewBufferString(tt.input)}
				var buf bytes.Buffer
				chooseStdout = mockWriteCloser{Writer: &buf}

				defer func() {
					chooseStdin = nil
					chooseStdout = nil
				}()
			}

			result, err := ChooseOne("テストプロンプト", tt.list)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ChooseOne() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.expectedErr {
				t.Errorf("ChooseOne() error message = %v, expectedErr %v", err.Error(), tt.expectedErr)
			}
			if result != tt.wantResult {
				t.Errorf("ChooseOne() result = %v, wantResult %v", result, tt.wantResult)
			}
		})
	}
}

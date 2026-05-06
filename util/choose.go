// Package util はアプリケーション全体で共有されるユーティリティ関数を提供します。
package util

import (
	"errors"
	"io"

	"github.com/manifoldco/promptui"
)

var (
	// テスト時に標準入力・標準出力をモックするための変数
	chooseStdin  io.ReadCloser
	chooseStdout io.WriteCloser
)

// ChooseOne は引数として受け取った文字列スライスをターミナル上に表示し、
// ユーザーが矢印キーで選択した要素を返します。
func ChooseOne(prompt string, list []string) (string, error) {
	if len(list) == 0 {
		Logger.Error(
			"choose.go: ChooseOne()",
			"The list is empty",
			"リストが空です",
		)
		err := errors.New("the list is empty")
		return "", err
	}

	selector := promptui.Select{
		Label:  prompt,
		Items:  list,
		Stdin:  chooseStdin,
		Stdout: chooseStdout,
	}

	_, result, err := selector.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

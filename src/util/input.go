package util

import "github.com/AlecAivazis/survey/v2"

// Input は、プロンプトメッセージ msg を出力し、標準入力から入力された文字列を1行読み込んで返します。
// 戻り値として、入力された文字列と発生したエラーを返します。
func Input(msg string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: msg,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

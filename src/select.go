// Package src はプログラムを構成する部品に分割して格納する
package src

import "github.com/AlecAivazis/survey/v2"

// Select はプロンプト用のメッセージと選択肢の配列を受け取り、選択された選択肢を返す
func Select(message string, options []string) (selectedOption string, err error) {
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err = survey.AskOne(prompt, &selectedOption)
	return selectedOption, err
}

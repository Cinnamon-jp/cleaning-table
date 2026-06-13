package util

import (
	"bufio"
	"fmt"
	"os"
)

// Input は、プロンプトメッセージ msg を出力し、標準入力から入力された文字列を1行読み込んで返します。
func Input(msg string) string {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

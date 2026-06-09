// Package src はプログラムを構成する部品に分割して格納する
package src

// ConvertExcel はExcelから取り出したデータを処理しやすいように変換する
func ConvertExcel(excelData [][]string) (convertedData [][]string, err error) {
	// Excelデータの文法をチェックする
	var isOK bool
	if isOK, err = checkExcel(excelData); !isOK || err != nil {
		Logger(Error, "convertExcel.go/ConvertExcel()/checkExcel()", "Error when executing checkExcel()", "checkExcel()の実行中にエラーが発生しました")
		return nil, err
	}
	
	// コメント列を削除
	// 部屋番号範囲の展開


	return nil, nil
}

// checkExcel はExcelの文法チェックを行う 未完成
func checkExcel(excelData [][]string) (isOK bool, err error) {
	return true, nil
}

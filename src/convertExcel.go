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
	noCommentExcelData := removeComment(excelData)

	// 部屋番号範囲の展開


	return nil, nil
}

// checkExcel はExcelの文法チェックを行う 未完成
func checkExcel(excelData [][]string) (isOK bool, err error) {
	return true, nil
}

// removeComment はコメント行(1行目+1列目)を削除する
func removeComment(excelData [][]string) [][]string {
	// データが空、または1行しかない場合は空のスライスを返す
	if len(excelData) <= 1 {
		return [][]string{}
	}

	// 元のスライスに影響を与えないよう、新しいスライスを作成する
	var result [][]string
	// 1行目をスキップ (excelData[1:])
	for _, row := range excelData[1:] {
		// 列が1列以下しかない場合は空の行を追加
		if len(row) <= 1 {
			result = append(result, []string{})
		} else {
			// 1列目をスキップした要素をコピーして新しい行を作成
			newRow := make([]string, len(row)-1)
			copy(newRow, row[1:])
			result = append(result, newRow)
		}
	}

	return result
}
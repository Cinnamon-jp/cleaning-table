package main

import (
	"flag"
	"math/rand/v2"
	"os"
)

func main() {
	// コマンドライン引数の設定
	inputFile := flag.String("input", "test.xlsx", "入力するエクセルファイルのパス")
	outputFile := flag.String("output", "output.pdf", "出力するPDFファイルのパス")
	flag.Parse()

	if err := run(*inputFile, *outputFile); err != nil {
		LogError("Failed to run", "実行に失敗しました", "error", err)
		os.Exit(1)
	}
}

func run(inputFile, outputFile string) error {
	LogInfo("Starting cleaning table generation", "掃除当番表の生成を開始します")

	// 1. エクセルファイルの解析
	LogInfo("Parsing Excel file", "エクセルファイルの解析を行っています", "file", inputFile)
	parser := NewExcelParser()
	data, err := parser.ParseExcelFile(inputFile)
	if err != nil {
		return err
	}
	LogInfo("Parsed Excel file successfully", "エクセルファイルの解析に成功しました", "columns", len(data.Columns))

	// 2. 掃除当番のシャッフル・割り当て
	LogInfo("Shuffling and assigning roles", "掃除当番のシャッフルと割り当てを行っています")
	// ランダムシード付きの乱数生成器を使用（毎回異なる結果になるように）
	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	shuffleResult := ShuffleAssign(data, r)
	LogInfo("Shuffled roles successfully", "役職のシャッフルが完了しました")

	// 3. PDFファイルの生成
	LogInfo("Generating PDF file", "PDFファイルの生成を行っています", "output", outputFile)
	err = GeneratePDF(shuffleResult, outputFile)
	if err != nil {
		return err
	}
	LogInfo("PDF generated successfully", "PDFファイルの生成に成功しました")

	return nil
}

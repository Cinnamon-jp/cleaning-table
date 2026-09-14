// Package main は掃除当番表生成ツールのCLIエントリポイントです。
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"cleaning-table/internal/config"
	"cleaning-table/internal/matcher"
	"cleaning-table/internal/output"
	"cleaning-table/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "ideal_test.yaml", "設定ファイルのパス (YAML)")
	historyPath := flag.String("history", "history.json", "過去担当履歴ファイルのパス (JSON)")
	pdfPath := flag.String("pdf", "cleaning_table.pdf", "出力するPDF当番表のパス (空でスキップ)")
	txtPath := flag.String("txt", "reverse_table.txt", "出力する逆引きテキストのパス (空でスキップ)")
	fontPath := flag.String("font", "NotoSerifJP-VariableFont_wght.ttf", "日本語 TrueType フォントファイルのパス")
	flag.Parse()

	// 1. 設定ファイルの読み込み
	slog.Info("設定ファイルを読み込んでいます...", "config", *configPath)
	rules, err := config.LoadFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}
	slog.Info("設定ファイルを読み込みました", "rules_count", len(rules))

	// 2. 過去履歴の読み込み
	slog.Info("過去の担当履歴を読み込んでいます...", "history", *historyPath)
	history, err := storage.LoadHistory(*historyPath)
	if err != nil {
		return fmt.Errorf("履歴ファイルの読み込みに失敗しました: %w", err)
	}
	slog.Info("過去履歴を読み込みました", "recorded_rooms", len(history))

	// 3. 公平割り当ての実行（ハンガリアン法）
	slog.Info("掃除当番の公平割り当て（シャッフル）を実行しています...")
	results, newHistory, err := matcher.ShuffleTask(rules, history)
	if err != nil {
		return fmt.Errorf("タスク割り当ての実行に失敗しました: %w", err)
	}
	slog.Info("割り当てが完了しました", "assigned_count", len(results))

	// 4. 逆引きテキストの出力
	if *txtPath != "" {
		slog.Info("逆引きテキストを出力しています...", "path", *txtPath)
		if err := output.WriteReverseText(results, *txtPath); err != nil {
			return fmt.Errorf("逆引きテキストの出力に失敗しました: %w", err)
		}
		slog.Info("逆引きテキストを出力しました", "path", *txtPath)
	}

	// 5. PDF 当番表の出力
	if *pdfPath != "" {
		slog.Info("PDF当番表を出力しています...", "path", *pdfPath, "font", *fontPath)
		if err := output.WritePDF(results, *fontPath, *pdfPath); err != nil {
			return fmt.Errorf("PDF当番表の出力に失敗しました: %w", err)
		}
		slog.Info("PDF当番表を出力しました", "path", *pdfPath)
	}

	// 6. 更新された履歴の保存
	if *historyPath != "" {
		slog.Info("更新された履歴を保存しています...", "path", *historyPath)
		if err := storage.SaveHistory(newHistory, *historyPath); err != nil {
			return fmt.Errorf("履歴ファイルの保存に失敗しました: %w", err)
		}
		slog.Info("履歴ファイルを更新しました", "path", *historyPath)
	}

	slog.Info("すべての処理が正常に完了しました！")
	return nil
}

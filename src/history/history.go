// Package history は掃除当番の割り当て履歴の永続化を担当する
package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	"cleaning-table/src/model"
)

// LoadHistory は指定されたパスからJSON形式の履歴ファイルを読み込む。
// ファイルが存在しない場合は空のHistoryを返す。
func LoadHistory(path string) (*model.History, error) {
	data, err := os.ReadFile(path) //nolint:gosec // 履歴ファイルのパスはプログラム内で固定されており安全
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			slog.Info("history file not found, starting fresh / 履歴ファイルが見つかりません。新規作成します", slog.String("path", path))
			return &model.History{
				Counts: make(map[string]map[string]int),
			}, nil
		}
		return nil, fmt.Errorf("reading history file / 履歴ファイルの読み込み中: %w", err)
	}

	var history model.History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("parsing history file / 履歴ファイルの解析中: %w", err)
	}

	if history.Counts == nil {
		history.Counts = make(map[string]map[string]int)
	}

	return &history, nil
}

// SaveHistory は履歴をJSON形式で指定されたパスに保存する。
func SaveHistory(path string, history *model.History) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling history / 履歴のJSON変換中: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing history file / 履歴ファイルの書き込み中: %w", err)
	}

	slog.Info("history file saved / 履歴ファイルを保存しました", slog.String("path", path))
	return nil
}

// GenerateFingerprint はPostSetのスライスから、部屋番号と役職名の構成を表すフィンガープリント文字列を生成する。
// Excelの構成変更を検知するために使用する。
func GenerateFingerprint(postSets []model.PostSet) string {
	var parts []string

	for _, ps := range postSets {
		// 部屋番号をソートして文字列化
		sortedRooms := make([]int, len(ps.RoomNumbers))
		copy(sortedRooms, ps.RoomNumbers)
		sort.Ints(sortedRooms)

		roomStrs := make([]string, len(sortedRooms))
		for i, r := range sortedRooms {
			roomStrs[i] = strconv.Itoa(r)
		}

		// 役職名をソートして文字列化
		sortedPosts := make([]string, len(ps.Posts))
		copy(sortedPosts, ps.Posts)
		sort.Strings(sortedPosts)

		part := strings.Join(roomStrs, ",") + ":" + strings.Join(sortedPosts, ",")
		parts = append(parts, part)
	}

	// PostSet間の順序も正規化
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

// CheckAndResetHistory はフィンガープリントを比較し、不一致の場合は履歴をリセットする。
// 不一致の場合はtrueを返す。
func CheckAndResetHistory(history *model.History, newFingerprint string) bool {
	if history.SourceFingerprint == newFingerprint {
		return false
	}

	if history.SourceFingerprint != "" {
		slog.Warn("Excel configuration changed, resetting history / Excelの構成が変更されたため、履歴をリセットします")
	}

	history.SourceFingerprint = newFingerprint
	history.Counts = make(map[string]map[string]int)
	return true
}

// UpdateCounts は割り当て結果で累計カウントを更新する。
func UpdateCounts(history *model.History, results []model.ShuffledPostSet) {
	for _, result := range results {
		roomKey := strconv.Itoa(result.RoomNumber)

		if history.Counts[roomKey] == nil {
			history.Counts[roomKey] = make(map[string]int)
		}

		history.Counts[roomKey][result.Post]++
	}
}

// GetCountMap は履歴から部屋番号(int)をキーとするカウントマップを構築する。
func GetCountMap(history *model.History) map[int]map[string]int {
	countMap := make(map[int]map[string]int)

	for roomKey, postCounts := range history.Counts {
		roomNum, err := strconv.Atoi(roomKey)
		if err != nil {
			continue
		}
		countMap[roomNum] = postCounts
	}

	return countMap
}

// Package storage は過去の担当履歴等の永続化を提供します。
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cleaning-table/internal/domain"
)

// LoadHistory は指定されたパスの JSON ファイルから AssignHistory を読み込みます。
// ファイルが存在しない場合は空の AssignHistory を返します。
func LoadHistory(path string) (domain.AssignHistory, error) {
	cleanPath := filepath.Clean(path)
	//nolint:gosec // G304: ユーザー指定の履歴ファイルパスを読み込むため
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(domain.AssignHistory), nil
		}
		return nil, fmt.Errorf("failed to read history file %q: %w", path, err)
	}

	if len(data) == 0 {
		return make(domain.AssignHistory), nil
	}

	var history domain.AssignHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history JSON from %q: %w", path, err)
	}

	if history == nil {
		history = make(domain.AssignHistory)
	}

	return history, nil
}

// SaveHistory は AssignHistory をインデント付き JSON 形式で指定されたパスに保存します。
func SaveHistory(history domain.AssignHistory, path string) error {
	cleanPath := filepath.Clean(path)

	dir := filepath.Dir(cleanPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history to JSON: %w", err)
	}

	//nolint:gosec // G306: 履歴ファイルは通常の読み書き権限で作成
	if err := os.WriteFile(cleanPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write history file %q: %w", path, err)
	}

	return nil
}

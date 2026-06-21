// Package model はプログラムで使用するデータモデルの型を定義する
package model

// History は過去の掃除当番の割り当て累計カウントを保持する。
// SourceFingerprint はExcelの部屋番号+役職構成のハッシュで、
// Excelが変更された場合の検知に使用する。
type History struct {
	SourceFingerprint string                    `json:"source_fingerprint"`
	Counts            map[string]map[string]int `json:"counts"`
}

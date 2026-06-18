// Package model はプログラムで使用するデータモデルの型を定義する
package model

// PostSet はExcelから読み込んだ部屋番号群と役職名・役職数のセットを表す。
type PostSet struct {
	RoomNumbers []int
	Posts       []string
	PostCounts  []int
}

// ShuffledPostSet はシャッフル後の部屋番号と割り振られた掃除場所のペアを表す。
type ShuffledPostSet struct {
	RoomNumber int
	Post       string
}

// Package model はプログラムで使用するデータモデルの型を定義する
package model

type PostSet struct {
	RoomNumbers []int
	Posts       []string
	PostCounts  []int
}

type ShuffledPostSet struct {
	RoomNumber int
	Post       string
}

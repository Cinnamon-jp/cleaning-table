// Package model はプログラムで使用するデータモデルの型を定義する
package model

type PostSet struct {
	RoomNumbers []int
	Posts       []string
	PostCounts  []int
}

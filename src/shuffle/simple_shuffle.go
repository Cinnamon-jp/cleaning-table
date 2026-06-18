// Package shuffle は役職をシャッフルする処理を格納する
package shuffle

import (
	"errors"
	"math/rand/v2"
	"slices"

	"cleaning-table/src/model"
)

// SimpleShuffle は各PostSetの部屋番号をシャッフルし、各部屋に掃除場所を割り振ったShuffledPostSetのスライスを返す。
func SimpleShuffle(postSets []model.PostSet) ([]model.ShuffledPostSet, error) {
	var returnPostSet []model.ShuffledPostSet

	for _, postSet := range postSets {
		// 役職名の数と役職数の数が一致しているか確認
		if len(postSet.Posts) != len(postSet.PostCounts) {
			return nil, errors.New("mismatch between posts and post counts / 役職名と役職数の不一致")
		}

		// 役職名と役職数を用いて、役職を展開
		var posts []string
		for postIdx, post := range postSet.Posts {
			repeatPosts := slices.Repeat([]string{post}, postSet.PostCounts[postIdx])
			posts = append(posts, repeatPosts...)
		}

		// 部屋番号をシャッフル
		shuffledRoomNumbers := shuffleSlice(postSet.RoomNumbers)

		// 数の不一致エラーを処理
		if len(shuffledRoomNumbers) != len(posts) {
			return nil, errors.New("mismatch between room numbers and posts / 部屋番号と役職の不一致")
		}

		// 最終的な変数に代入
		for i, roomNumber := range shuffledRoomNumbers {
			returnPostSet = append(returnPostSet, model.ShuffledPostSet{
				RoomNumber: roomNumber,
				Post:       posts[i],
			})
		}
	}

	return returnPostSet, nil
}

// shuffleSlice は受け取ったスライスをシャッフルし、新しいスライスを返す
func shuffleSlice[T any](slice []T) []T {
	// 元のスライスをコピーして元データを保護する
	result := make([]T, len(slice))
	copy(result, slice)

	// Fisher-Yatesアルゴリズムでシャッフル
	for i := len(result) - 1; i > 0; i-- {
		j := rand.IntN(i + 1) //nolint:gosec // セキュリティ用途ではなく掃除当番のシャッフルのため暗号論的乱数は不要
		result[i], result[j] = result[j], result[i]
	}

	return result
}

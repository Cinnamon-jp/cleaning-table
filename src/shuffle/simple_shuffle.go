// Package shuffle は役職をシャッフルする処理を格納する
package shuffle

import (
	"cleaning-table/src/model"
	"cleaning-table/src/util"
	"errors"
	"math/rand"
	"slices"
)

func SimpleShuffle(postSets []model.PostSet) ([]model.ShuffledPostSet, error) {
	var returnPostSet []model.ShuffledPostSet

	for _, postSet := range postSets {
		// 役職名の数と役職数の数が一致しているか確認
		if len(postSet.Posts) != len(postSet.PostCounts) {
			util.Logger(util.Error, "simple_shuffle.go/SimpleShuffle()/len(postSet.Posts) != len(postSet.PostCounts)", "Don't match posts and post counts", "役職名の数と役職数の数が一致しません")
			return nil, errors.New("役職名の数と役職数の数が一致しません")
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
			util.Logger(util.Error, "simple_shuffle.go/SimpleShuffle()/len(shuffledRoomNumbers) != len(posts)", "Don't match room numbers and posts", "部屋番号の数と役職数が一致しません")
			return nil, errors.New("部屋番号の数と役職の数が一致しません")
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
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
}

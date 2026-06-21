// Package shuffle は役職をシャッフルする処理を格納する
package shuffle

import (
	"errors"
	"math/rand/v2"
	"slices"

	"cleaning-table/src/model"
)

// EvenShuffle は過去の履歴を考慮して、役職の偏りが少なくなるように重み付きランダムで割り当てを行う。
// countMap は部屋番号→役職名→累計割り当て回数のマップ。nilの場合は完全ランダムと同等に動作する。
func EvenShuffle(postSets []model.PostSet, countMap map[int]map[string]int) ([]model.ShuffledPostSet, error) {
	if countMap == nil {
		countMap = make(map[int]map[string]int)
	}

	var returnPostSet []model.ShuffledPostSet

	for _, postSet := range postSets {
		// 役職名の数と役職数の数が一致しているか確認
		if len(postSet.Posts) != len(postSet.PostCounts) {
			return nil, errors.New("mismatch between posts and post counts / 役職名と役職数の不一致")
		}

		// 残数マップを作成
		remaining := make(map[string]int)
		for i, post := range postSet.Posts {
			remaining[post] += postSet.PostCounts[i]
		}

		// 展開済み役職の総数を計算
		totalPosts := 0
		for _, count := range remaining {
			totalPosts += count
		}

		// 数の不一致エラーを処理
		if len(postSet.RoomNumbers) != totalPosts {
			return nil, errors.New("mismatch between room numbers and posts / 部屋番号と役職の不一致")
		}

		// 部屋番号をシャッフル（処理順のランダム化）
		shuffledRooms := shuffleSlice(postSet.RoomNumbers)

		// 各部屋に重み付きランダムで役職を割り当てる
		for _, room := range shuffledRooms {
			post, err := weightedSelect(room, remaining, countMap)
			if err != nil {
				return nil, err
			}

			returnPostSet = append(returnPostSet, model.ShuffledPostSet{
				RoomNumber: room,
				Post:       post,
			})

			remaining[post]--
			if remaining[post] == 0 {
				delete(remaining, post)
			}
		}
	}

	return returnPostSet, nil
}

// calcWeight は累計割り当て回数から重みを計算する。
// w = 1 / (1 + count) で、過去に多く割り当てられた役職ほど重みが小さくなる。
func calcWeight(count int) float64 {
	return 1.0 / (1.0 + float64(count))
}

// weightedSelect は残数マップと過去の履歴に基づいて、重み付きランダムで1つの役職を選択する。
func weightedSelect(room int, remaining map[string]int, countMap map[int]map[string]int) (string, error) {
	roomCounts := countMap[room]

	// 各候補役職のスコアを計算
	type candidate struct {
		post  string
		score float64
	}

	var candidates []candidate
	totalScore := 0.0

	for post, rem := range remaining {
		if rem <= 0 {
			continue
		}

		count := 0
		if roomCounts != nil {
			count = roomCounts[post]
		}

		score := calcWeight(count) * float64(rem)
		candidates = append(candidates, candidate{post: post, score: score})
		totalScore += score
	}

	if len(candidates) == 0 {
		return "", errors.New("no available posts to assign / 割り当て可能な役職がありません")
	}

	// 重み付きランダム選択
	r := rand.Float64() * totalScore //nolint:gosec // セキュリティ用途ではなく掃除当番のシャッフルのため暗号論的乱数は不要
	cumulative := 0.0
	for _, c := range candidates {
		cumulative += c.score
		if r <= cumulative {
			return c.post, nil
		}
	}

	// 浮動小数点の丸め誤差対策として最後の候補を返す
	return candidates[len(candidates)-1].post, nil
}

// shuffleSlice は受け取ったスライスをシャッフルし、新しいスライスを返す
func shuffleSlice[T any](slice []T) []T {
	// 元のスライスをコピーして元データを保護する
	result := slices.Clone(slice)

	// Fisher-Yatesアルゴリズムでシャッフル
	for i := len(result) - 1; i > 0; i-- {
		j := rand.IntN(i + 1) //nolint:gosec // セキュリティ用途ではなく掃除当番のシャッフルのため暗号論的乱数は不要
		result[i], result[j] = result[j], result[i]
	}

	return result
}

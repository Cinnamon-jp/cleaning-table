// Package validator は割り当てルールおよび部屋のカバー率の検証を行います。
package validator

import (
	"fmt"

	"cleaning-table/internal/domain"
)

// CheckDuplicatesAndSparse はルール集合の中で最大の要素数を持つルール（全部屋ルール）から
// 他のルールの部屋集合を差し引き、すべての部屋が過不足なくカバーされているかを検証します。
func CheckDuplicatesAndSparse(rules []domain.AssignRule) error {
	if len(rules) == 0 {
		return nil
	}

	maxIdx := 0 // 要素数が最大のインデックス（全部屋が入っていることとする）
	maxN := 0
	for i, rule := range rules {
		if rule.Rooms == nil {
			continue
		}
		nRooms := rule.Rooms.Cardinality()
		if maxN < nRooms {
			maxIdx = i
			maxN = nRooms
		}
	}

	if rules[maxIdx].Rooms == nil {
		return nil
	}

	// 全部屋が入った要素だけ抜いたインデックス
	targetIdx := make([]int, 0, len(rules)-1)
	for i := range rules {
		if i == maxIdx {
			continue
		}
		targetIdx = append(targetIdx, i)
	}

	// 元のデータを破壊しないように Clone して差集合を計算
	allRoom := rules[maxIdx].Rooms.Clone()
	for _, i := range targetIdx {
		if rules[i].Rooms != nil {
			allRoom = allRoom.Difference(rules[i].Rooms)
		}
	}

	// 最終的に空集合になっているか確認
	if !allRoom.IsEmpty() {
		return fmt.Errorf("rooms are not fully covered, remaining: %v", allRoom)
	}

	return nil
}

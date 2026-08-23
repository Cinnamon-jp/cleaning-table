package main

import (
	"fmt"
)

func checkDuplicatesAndSparse(rules []assignRule) error {
	if len(rules) == 0 {
		return nil
	}

	maxIdx := 0 // 要素数が最大のインデックス（全部屋が入っていることとする）
	maxN := 0
	for i, rule := range rules {
		if rule.rooms == nil {
			continue
		}
		nRooms := rule.rooms.Cardinality()
		if maxN < nRooms {
			maxIdx = i
			maxN = nRooms
		}
	}

	if rules[maxIdx].rooms == nil {
		return nil
	}

	// 全部屋が入った要素だけ抜いたインデックス
	targetIdx := make([]int, 0, len(rules)-1)
	for i := 0; i < len(rules); i++ {
		if i == maxIdx {
			continue
		}
		targetIdx = append(targetIdx, i)
	}

	// 元のデータを破壊しないように Clone して差集合を計算
	allRoom := rules[maxIdx].rooms.Clone()
	for _, i := range targetIdx {
		if rules[i].rooms != nil {
			allRoom = allRoom.Difference(rules[i].rooms)
		}
	}

	// 最終的に空集合になっているか確認
	if !allRoom.IsEmpty() {
		return fmt.Errorf("rooms are not fully covered, remaining: %v", allRoom)
	}

	return nil
}


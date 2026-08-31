// Package validator は割り当てルールおよび部屋のカバー率の検証を行います。
package validator

import (
	"fmt"
	"slices"

	"cleaning-table/internal/domain"
)

// CheckDuplicatesAndSparse は各部屋が過不足なく（重複なくちょうど1回ずつ）カバーされているかを検証します。
// ルール集合の中で最大の部屋数を持つ要素を「全部屋ルール」とみなし、
// それ以外の部分ルールによってすべての部屋がちょうど1回ずつ割り当てられているかを確認します。
func CheckDuplicatesAndSparse(rules []domain.AssignRule) error {
	if len(rules) == 0 {
		return nil
	}

	maxIdx := 0
	maxN := 0
	for i, rule := range rules {
		nRooms := len(rule.Rooms)
		if maxN < nRooms {
			maxIdx = i
			maxN = nRooms
		}
	}

	allRooms := rules[maxIdx].Rooms
	if len(allRooms) == 0 {
		return nil
	}

	// 1. 全部屋ルール内に重複がないか確認
	allRoomsMap := make(map[int]bool, len(allRooms))
	for _, room := range allRooms {
		if allRoomsMap[room] {
			return fmt.Errorf("all-rooms rule contains duplicate room: %d", room)
		}
		allRoomsMap[room] = true
	}

	// 2. 部分ルールの存在確認
	hasSubRules := false
	assignedCounts := make(map[int]int, len(allRooms))

	for i, rule := range rules {
		if i == maxIdx {
			continue
		}
		hasSubRules = true

		// 各部分ルール内の自己重複チェック
		ruleSeen := make(map[int]bool, len(rule.Rooms))
		for _, room := range rule.Rooms {
			if ruleSeen[room] {
				return fmt.Errorf("rule[%d] contains duplicate room: %d", i, room)
			}
			ruleSeen[room] = true

			// 全体ルールに含まれない部屋番号の検知
			if !allRoomsMap[room] {
				return fmt.Errorf("rule[%d] contains unknown room %d not in all-rooms rule", i, room)
			}

			assignedCounts[room]++
		}
	}

	if !hasSubRules {
		return fmt.Errorf("no sub-rules provided to cover all rooms")
	}

	// 3. 過不足（未カバーまたは重複割り当て）の検証
	var missing []int
	var duplicates []int

	for _, room := range allRooms {
		count := assignedCounts[room]
		if count == 0 {
			missing = append(missing, room)
		} else if count > 1 {
			duplicates = append(duplicates, room)
		}
	}

	if len(missing) > 0 || len(duplicates) > 0 {
		slices.Sort(missing)
		slices.Sort(duplicates)
		if len(missing) > 0 && len(duplicates) > 0 {
			return fmt.Errorf("rooms are not fully covered: missing %v, duplicated %v", missing, duplicates)
		}
		if len(missing) > 0 {
			return fmt.Errorf("rooms are not fully covered, remaining: %v", missing)
		}
		return fmt.Errorf("rooms have duplicate assignments: %v", duplicates)
	}

	return nil
}

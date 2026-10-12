// Package config は掃除当番表の設定読み込みおよび解釈を提供します。
package config

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config は掃除当番表の全体設定を保持します。
type Config struct {
	Groups []Group `yaml:"-"`
}

// Group は清掃の割り当て単位（部屋リストおよびタスク一覧）を表します。
type Group struct {
	Name  string
	Rooms []uint
	Tasks []Task

	// 式の保存用
	rawRoomsExpr string
}

// Task はグループ内で割り当てる清掃タスクを表します。
type Task struct {
	Name   string
	Number uint
	IsAuto bool
}

// UnmarshalYAML は YAML ノードから設定を動的にパースし、式や auto タスクを解決します。
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("root element must be mapping")
	}

	var groups []Group

	// トップレベルの全要素を定義順に走査
	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valNode := value.Content[i+1]

		if err := parseNodeRecursively(keyNode.Value, valNode, &groups); err != nil {
			return err
		}
	}

	// 式の解決と auto の計算
	if err := resolveRoomExpressions(groups); err != nil {
		return err
	}

	c.Groups = groups
	return nil
}

// parseNodeRecursively はノードを再帰的に走査して Group を収集します。
func parseNodeRecursively(currentPath string, node *yaml.Node, out *[]Group) error {
	switch node.Kind {
	case yaml.SequenceNode:
		// 配列形式（例: facility: [101, 102] や all_rooms.1F: [101, ...]）
		var rooms []uint
		if err := node.Decode(&rooms); err != nil {
			return fmt.Errorf("%s fail to decode room number: %w", currentPath, err)
		}
		*out = append(*out, Group{
			Name:  currentPath,
			Rooms: rooms,
		})
		return nil

	case yaml.MappingNode:
		// "rooms" または "tasks" を含む場合は末端のグループ定義
		if hasKey(node, "rooms") || hasKey(node, "tasks") {
			grp, err := parseGroupLeaf(currentPath, node)
			if err != nil {
				return err
			}
			*out = append(*out, grp)
			return nil
		}

		// それ以外は中間階層（例: all_rooms, posts, common）
		for i := 0; i < len(node.Content); i += 2 {
			subKey := node.Content[i].Value
			subVal := node.Content[i+1]
			nextPath := currentPath + "." + subKey

			if err := parseNodeRecursively(nextPath, subVal, out); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("undefined node style: path=%s (kind=%d)", currentPath, node.Kind)
	}
}

// parseGroupLeaf は rooms / tasks を持つグループ末端をパースします。
func parseGroupLeaf(name string, node *yaml.Node) (Group, error) {
	grp := Group{Name: name}

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		val := node.Content[i+1]

		switch key {
		case "rooms":
			switch val.Kind {
			case yaml.SequenceNode:
				if err := val.Decode(&grp.Rooms); err != nil {
					return Group{}, fmt.Errorf("fail to decode %s: %w", name, err)
				}
			case yaml.ScalarNode:
				grp.rawRoomsExpr = val.Value
			default:
				return Group{}, fmt.Errorf("%s's rooms must be slice or string", name)
			}

		case "tasks":
			if val.Kind != yaml.MappingNode {
				return Group{}, fmt.Errorf("%s's tasks must be mapping", name)
			}
			for j := 0; j < len(val.Content); j += 2 {
				taskName := val.Content[j].Value
				taskVal := val.Content[j+1]

				t := Task{Name: taskName}
				if taskVal.Value == "auto" {
					t.IsAuto = true
				} else {
					var num uint
					if err := taskVal.Decode(&num); err != nil {
						return Group{}, fmt.Errorf("%s のタスク %s の数値デコードに失敗しました: %w", name, taskName, err)
					}
					t.Number = num
				}
				grp.Tasks = append(grp.Tasks, t)
			}
		}
	}

	return grp, nil
}

// hasKey() returns true if MappingNode has specified key
func hasKey(node *yaml.Node, key string) bool {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return true
		}
	}
	return false
}

// resolveRoomExpressions は集合演算式を解決し、auto タスクの人数を算出します。
func resolveRoomExpressions(groups []Group) error {
	// 式の解決
	for i := range groups {
		if groups[i].rawRoomsExpr == "" {
			continue
		}

		expr := groups[i].rawRoomsExpr
		tokens := strings.Split(expr, "-")
		if len(tokens) == 0 {
			continue
		}

		firstToken := strings.TrimSpace(tokens[0])
		baseRooms := findRooms(groups, firstToken)
		if len(baseRooms) == 0 {
			return fmt.Errorf("グループ %s の式 '%s': 参照先 '%s' に部屋が存在しません", groups[i].Name, expr, firstToken)
		}

		// 差集合を計算
		roomSet := make(map[uint]bool, len(baseRooms))
		for _, r := range baseRooms {
			roomSet[r] = true
		}

		for _, token := range tokens[1:] {
			subToken := strings.TrimSpace(token)
			if subToken == "" {
				continue
			}
			removeRooms := findRooms(groups, subToken)
			for _, r := range removeRooms {
				delete(roomSet, r)
			}
		}

		// 元の baseRooms の順序を保持して残った部屋を格納
		var result []uint
		for _, r := range baseRooms {
			if roomSet[r] {
				result = append(result, r)
			}
		}

		groups[i].Rooms = result
	}

	// auto タスクの人数の自動計算 (部屋数 - 他の固定タスク合計)
	for i := range groups {
		var fixedTotal uint
		var autoTasks []*Task
		for j := range groups[i].Tasks {
			if groups[i].Tasks[j].IsAuto {
				autoTasks = append(autoTasks, &groups[i].Tasks[j])
			} else {
				fixedTotal += groups[i].Tasks[j].Number
			}
		}
		if len(autoTasks) > 0 {
			roomCount := uint(len(groups[i].Rooms))
			if roomCount >= fixedTotal {
				remaining := roomCount - fixedTotal
				perTask := remaining / uint(len(autoTasks))
				for _, at := range autoTasks {
					at.Number = perTask
				}
			}
		}
	}

	return nil
}

// findRooms はグループ名（完全一致またはプレフィックス前方一致）から部屋一覧を取得します。
func findRooms(groups []Group, identifier string) []uint {
	var result []uint
	foundExact := false

	// 完全一致（例: "all_rooms.1F" や "facility"）
	for _, g := range groups {
		if g.Name == identifier {
			result = append(result, g.Rooms...)
			foundExact = true
			break
		}
	}

	// 完全一致がない場合は前方一致（例: "posts" -> "posts.leader", "posts.mentor" など全て）
	if !foundExact {
		prefix := identifier + "."
		for _, g := range groups {
			if strings.HasPrefix(g.Name, prefix) {
				result = append(result, g.Rooms...)
			}
		}
	}

	return result
}

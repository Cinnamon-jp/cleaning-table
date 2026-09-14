// Package config は YAML 設定ファイルの読み込み、シンボル解決、式評価、および AssignRule の生成を提供します。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"cleaning-table/internal/domain"
	"cleaning-table/internal/expr"
)

// LoadFromFile は指定されたパスの YAML 設定ファイルを読み込み、AssignRule のスライスを生成します。
func LoadFromFile(path string) ([]domain.AssignRule, error) {
	//nolint:gosec // G304: ユーザーから指定された設定ファイルパスを読み込むため
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}
	return LoadFromBytes(data)
}

// LoadFromBytes は YAML バイト列から AssignRule のスライスを生成します。
func LoadFromBytes(data []byte) ([]domain.AssignRule, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(root.Content) == 0 {
		return nil, nil
	}

	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("line %d: root of YAML must be a mapping, got %v", doc.Line, doc.Kind)
	}

	symbols := make(map[string]expr.Set)
	var assignmentsNode *yaml.Node

	for i := 0; i < len(doc.Content); i += 2 {
		key := doc.Content[i].Value
		val := doc.Content[i+1]

		switch key {
		case "groups":
			if _, err := parseGroups(val, "groups", symbols); err != nil {
				return nil, err
			}
		case "assignments":
			assignmentsNode = val
		}
	}

	if assignmentsNode == nil {
		return nil, fmt.Errorf("missing 'assignments' section in YAML")
	}

	return parseAssignments(assignmentsNode, symbols)
}

// parseGroups は groups ノードを再帰的に走査し、各パスおよび親グループの和集合を symbols に登録します。
func parseGroups(node *yaml.Node, path string, symbols map[string]expr.Set) (expr.Set, error) {
	switch node.Kind {
	case yaml.SequenceNode:
		set := make(expr.Set, len(node.Content))
		for _, item := range node.Content {
			room, err := strconv.Atoi(item.Value)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid room number %q: %w", item.Line, item.Value, err)
			}
			set[room] = struct{}{}
		}
		registerSymbol(symbols, path, set)
		return set, nil

	case yaml.MappingNode:
		aggregated := make(expr.Set)
		for i := 0; i < len(node.Content); i += 2 {
			subKey := node.Content[i].Value
			subVal := node.Content[i+1]

			subPath := subKey
			if path != "" && path != "groups" {
				subPath = path + "." + subKey
			}

			subSet, err := parseGroups(subVal, subPath, symbols)
			if err != nil {
				return nil, err
			}
			aggregated = expr.Union(aggregated, subSet)
		}
		if path != "" && path != "groups" {
			registerSymbol(symbols, path, aggregated)
		}
		return aggregated, nil

	default:
		return nil, fmt.Errorf("line %d: unexpected node kind in groups: %v", node.Line, node.Kind)
	}
}

func registerSymbol(symbols map[string]expr.Set, path string, set expr.Set) {
	symbols[path] = set
	cleanPath := strings.TrimPrefix(path, "groups.")
	cleanPath = strings.TrimPrefix(cleanPath, "assignments.")
	symbols[cleanPath] = set
}

type rawRule struct {
	path      string
	roomsNode *yaml.Node
	tasksNode *yaml.Node
}

// collectRules は assignments ノードを再帰的に走査し、末端の割り当てルール一覧を収集します。
func collectRules(node *yaml.Node, currentPath string, rules *[]rawRule) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: expected mapping node, got %v", node.Line, node.Kind)
	}

	var roomsNode, tasksNode *yaml.Node
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		val := node.Content[i+1]
		switch key {
		case "rooms":
			roomsNode = val
		case "tasks":
			tasksNode = val
		}
	}

	// rooms と tasks の両方を持っていればルールとみなす
	if roomsNode != nil && tasksNode != nil {
		*rules = append(*rules, rawRule{
			path:      currentPath,
			roomsNode: roomsNode,
			tasksNode: tasksNode,
		})
		return nil
	}

	// 中間グループの場合は配下の全要素を再帰的に走査
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		val := node.Content[i+1]
		subPath := key
		if currentPath != "" {
			subPath = currentPath + "." + key
		}
		if err := collectRules(val, subPath, rules); err != nil {
			return err
		}
	}

	return nil
}

// parseAssignments は収集したルールを静的解決、式解決、タスク展開の順に処理します。
func parseAssignments(assignmentsNode *yaml.Node, symbols map[string]expr.Set) ([]domain.AssignRule, error) {
	var rawRules []rawRule
	if err := collectRules(assignmentsNode, "", &rawRules); err != nil {
		return nil, err
	}

	if len(rawRules) == 0 {
		return nil, fmt.Errorf("no assignment rules found in YAML")
	}

	// Pass 1: 静的な部屋リストを持つルールを解決 & シンボルテーブルに登録
	// 同時に親グループ（例: "posts"）へ部屋番号を集約
	for _, rule := range rawRules {
		if rule.roomsNode.Kind == yaml.SequenceNode {
			set := make(expr.Set, len(rule.roomsNode.Content))
			for _, item := range rule.roomsNode.Content {
				room, err := strconv.Atoi(item.Value)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid room number %q: %w", item.Line, item.Value, err)
				}
				set[room] = struct{}{}
			}
			registerSymbol(symbols, rule.path, set)

			// 親グループの集約（例: posts.leader -> posts）
			parts := strings.Split(rule.path, ".")
			for pIdx := 1; pIdx < len(parts); pIdx++ {
				parentPath := strings.Join(parts[:pIdx], ".")
				if existing, ok := symbols[parentPath]; ok {
					symbols[parentPath] = expr.Union(existing, set)
				} else {
					symbols[parentPath] = set.Clone()
				}
			}
		}
	}

	lookup := func(ident string) (expr.Set, error) {
		if s, ok := symbols[ident]; ok {
			return s, nil
		}
		if s, ok := symbols["groups."+ident]; ok {
			return s, nil
		}
		if s, ok := symbols["assignments."+ident]; ok {
			return s, nil
		}
		return nil, fmt.Errorf("undefined identifier %q", ident)
	}

	// Pass 2: 式を含むルールを評価 & シンボルテーブルに登録
	for _, rule := range rawRules {
		if rule.roomsNode.Kind == yaml.ScalarNode {
			exprStr := rule.roomsNode.Value
			set, err := expr.EvalToSet(exprStr, lookup)
			if err != nil {
				return nil, fmt.Errorf("line %d: failed to evaluate expression %q for rule %q: %w",
					rule.roomsNode.Line, exprStr, rule.path, err)
			}
			registerSymbol(symbols, rule.path, set)

			// 親グループの集約
			parts := strings.Split(rule.path, ".")
			for pIdx := 1; pIdx < len(parts); pIdx++ {
				parentPath := strings.Join(parts[:pIdx], ".")
				if existing, ok := symbols[parentPath]; ok {
					symbols[parentPath] = expr.Union(existing, set)
				} else {
					symbols[parentPath] = set.Clone()
				}
			}
		} else if rule.roomsNode.Kind != yaml.SequenceNode {
			return nil, fmt.Errorf("line %d: 'rooms' must be either a sequence or an expression string, got %v",
				rule.roomsNode.Line, rule.roomsNode.Kind)
		}
	}

	// Pass 3: 各ルールのタスク展開と AssignRule の生成
	assignRules := make([]domain.AssignRule, 0, len(rawRules))
	for _, rule := range rawRules {
		roomsSet, ok := symbols[rule.path]
		if !ok {
			return nil, fmt.Errorf("internal error: rooms for rule %q not found", rule.path)
		}
		rooms := roomsSet.ToSlice()

		tasks, err := expandTasks(rule.tasksNode, len(rooms), rule.path)
		if err != nil {
			return nil, err
		}

		assignRules = append(assignRules, domain.AssignRule{
			Rooms: rooms,
			Tasks: tasks,
		})
	}

	return assignRules, nil
}

type taskDef struct {
	name   string
	count  int
	isAuto bool
}

// expandTasks は tasks マッピングを解析し、部屋数に一致するタスク名のスライスに展開します。
func expandTasks(tasksNode *yaml.Node, roomCount int, rulePath string) ([]string, error) {
	if tasksNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("line %d: 'tasks' must be a mapping, got %v", tasksNode.Line, tasksNode.Kind)
	}

	var defs []taskDef
	fixedTotal := 0
	var autoTaskName string
	autoIdx := -1

	for i := 0; i < len(tasksNode.Content); i += 2 {
		taskName := tasksNode.Content[i].Value
		valNode := tasksNode.Content[i+1]

		valStr := strings.TrimSpace(valNode.Value)
		if valStr == "auto" || valStr == "?" {
			if autoTaskName != "" {
				return nil, fmt.Errorf("line %d: rule %q has multiple auto tasks: %q and %q",
					valNode.Line, rulePath, autoTaskName, taskName)
			}
			autoTaskName = taskName
			autoIdx = len(defs)
			defs = append(defs, taskDef{name: taskName, isAuto: true})
		} else {
			count, err := strconv.Atoi(valStr)
			if err != nil {
				return nil, fmt.Errorf("line %d: rule %q task %q has invalid count %q: %w",
					valNode.Line, rulePath, taskName, valStr, err)
			}
			if count < 0 {
				return nil, fmt.Errorf("line %d: rule %q task %q has negative count %d",
					valNode.Line, rulePath, taskName, count)
			}
			fixedTotal += count
			defs = append(defs, taskDef{name: taskName, count: count, isAuto: false})
		}
	}

	autoCount := roomCount - fixedTotal
	if autoCount < 0 {
		return nil, fmt.Errorf("line %d: rule %q: fixed task count (%d) exceeds room count (%d)",
			tasksNode.Line, rulePath, fixedTotal, roomCount)
	}

	if autoIdx != -1 {
		defs[autoIdx].count = autoCount
	} else if fixedTotal != roomCount {
		return nil, fmt.Errorf("line %d: rule %q: total task count (%d) does not match room count (%d) and no auto task specified",
			tasksNode.Line, rulePath, fixedTotal, roomCount)
	}

	tasks := make([]string, 0, roomCount)
	for _, def := range defs {
		for range def.count {
			tasks = append(tasks, def.name)
		}
	}

	return tasks, nil
}

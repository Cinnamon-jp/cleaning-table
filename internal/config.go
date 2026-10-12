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

// mapEntry はマッピングノードの 1 エントリ（エイリアス解決済み）を表します。
type mapEntry struct {
	key string
	val *yaml.Node
}

// UnmarshalYAML は YAML ノードから設定を動的にパースし、式や auto タスクを解決します。
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	value = deref(value)
	if value.Kind != yaml.MappingNode {
		return errors.New("root element must be mapping")
	}

	entries, err := mappingEntries("(root)", value)
	if err != nil {
		return err
	}

	var groups []Group

	// トップレベルの全要素を定義順に走査
	for _, e := range entries {
		if err := parseNodeRecursively(e.key, e.val, &groups); err != nil {
			return err
		}
	}

	// 同名グループ（例: "a.b" というキーと a: {b: ...}）の衝突を検出
	if err := checkDuplicateGroups(groups); err != nil {
		return err
	}

	// 式の解決と auto の計算
	if err := resolveRoomExpressions(groups); err != nil {
		return err
	}

	c.Groups = groups
	return nil
}

// parseNodeRecursively はノードを再帰的に走査して Group を収集します。
//
// "rooms" または "tasks" というキーを持つマッピングは末端のグループ定義として扱われるため、
// 中間階層のキー名としてこれらを使用することはできません。
func parseNodeRecursively(currentPath string, node *yaml.Node, out *[]Group) error {
	node = deref(node)

	if isNull(node) {
		return fmt.Errorf("%s の定義が空です", currentPath)
	}

	switch node.Kind {
	case yaml.SequenceNode:
		// 配列形式（例: facility: [101, 102] や all_rooms.1F: [101, ...]）
		rooms, err := decodeRooms(currentPath, node)
		if err != nil {
			return err
		}
		*out = append(*out, Group{
			Name:  currentPath,
			Rooms: rooms,
		})
		return nil

	case yaml.MappingNode:
		entries, err := mappingEntries(currentPath, node)
		if err != nil {
			return err
		}

		// "rooms" または "tasks" を含む場合は末端のグループ定義
		if hasKey(node, "rooms") || hasKey(node, "tasks") {
			grp, err := parseGroupLeaf(currentPath, entries)
			if err != nil {
				return err
			}
			*out = append(*out, grp)
			return nil
		}

		// それ以外は中間階層（例: all_rooms, posts, common）
		for _, e := range entries {
			if err := parseNodeRecursively(currentPath+"."+e.key, e.val, out); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("%s: 未対応の YAML 形式です (%s)", currentPath, kindName(node.Kind))
	}
}

// parseGroupLeaf は rooms / tasks を持つグループ末端をパースします。
func parseGroupLeaf(name string, entries []mapEntry) (Group, error) {
	grp := Group{Name: name}

	for _, e := range entries {
		switch e.key {
		case "rooms":
			if isNull(e.val) {
				return Group{}, fmt.Errorf("%s の rooms が空です", name)
			}
			switch e.val.Kind {
			case yaml.SequenceNode:
				rooms, err := decodeRooms(name, e.val)
				if err != nil {
					return Group{}, err
				}
				grp.Rooms = rooms
			case yaml.ScalarNode:
				// 数値などではなく文字列（式）のみを許可
				if e.val.ShortTag() != "!!str" || strings.TrimSpace(e.val.Value) == "" {
					return Group{}, fmt.Errorf("%s の rooms は配列または式(文字列)で指定してください", name)
				}
				grp.rawRoomsExpr = e.val.Value
			default:
				return Group{}, fmt.Errorf("%s の rooms は配列または式(文字列)で指定してください", name)
			}

		case "tasks":
			tasks, err := parseTasks(name, e.val)
			if err != nil {
				return Group{}, err
			}
			grp.Tasks = tasks

		default:
			return Group{}, fmt.Errorf("%s: 未知のキー '%s' があります (rooms / tasks のみ指定可能)", name, e.key)
		}
	}

	return grp, nil
}

// parseTasks は tasks マッピングをパースします。
func parseTasks(name string, node *yaml.Node) ([]Task, error) {
	if isNull(node) {
		return nil, fmt.Errorf("%s の tasks が空です", name)
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s の tasks はマッピングで指定してください", name)
	}

	entries, err := mappingEntries(name+".tasks", node)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for _, e := range entries {
		t := Task{Name: e.key}

		switch {
		case isNull(e.val):
			return nil, fmt.Errorf("%s のタスク %s の値が空です", name, e.key)
		case e.val.Kind != yaml.ScalarNode:
			return nil, fmt.Errorf("%s のタスク %s は数値または auto で指定してください", name, e.key)
		case e.val.ShortTag() == "!!str" && e.val.Value == "auto":
			t.IsAuto = true
		default:
			var num uint
			if err := e.val.Decode(&num); err != nil {
				return nil, fmt.Errorf("%s のタスク %s の数値デコードに失敗しました: %w", name, e.key, err)
			}
			t.Number = num
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// decodeRooms は部屋番号の配列をデコードし、重複を検出します。
func decodeRooms(name string, node *yaml.Node) ([]uint, error) {
	var rooms []uint
	if err := node.Decode(&rooms); err != nil {
		return nil, fmt.Errorf("%s の部屋番号のデコードに失敗しました: %w", name, err)
	}

	seen := make(map[uint]struct{}, len(rooms))
	for _, r := range rooms {
		if _, dup := seen[r]; dup {
			return nil, fmt.Errorf("%s の部屋番号 %d が重複しています", name, r)
		}
		seen[r] = struct{}{}
	}
	return rooms, nil
}

// mappingEntries はマッピングノードのエントリを定義順に返します。
// キーの重複、スカラー以外のキー、マージキー(<<)はエラーとし、値のエイリアスは解決します。
func mappingEntries(path string, node *yaml.Node) ([]mapEntry, error) {
	entries := make([]mapEntry, 0, len(node.Content)/2)
	seen := make(map[string]struct{}, len(node.Content)/2)

	for i := 0; i+1 < len(node.Content); i += 2 {
		k := node.Content[i]
		if k.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("%s: キーは文字列で指定してください (line %d)", path, k.Line)
		}
		if k.Tag == "!!merge" {
			return nil, fmt.Errorf("%s: マージキー(<<)は未対応です (line %d)", path, k.Line)
		}
		if _, dup := seen[k.Value]; dup {
			return nil, fmt.Errorf("%s: キー '%s' が重複しています (line %d)", path, k.Value, k.Line)
		}
		seen[k.Value] = struct{}{}

		entries = append(entries, mapEntry{key: k.Value, val: deref(node.Content[i+1])})
	}
	return entries, nil
}

// deref はエイリアスノードをアンカー先のノードに解決します。
func deref(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}
	return node
}

// isNull は node が null スカラーかどうかを返します。
func isNull(node *yaml.Node) bool {
	return node.Kind == yaml.ScalarNode && node.ShortTag() == "!!null"
}

// kindName は yaml.Kind を人が読める名前に変換します。
func kindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return fmt.Sprintf("kind=%d", k)
	}
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

// checkDuplicateGroups は同名のグループが存在しないことを検証します。
func checkDuplicateGroups(groups []Group) error {
	seen := make(map[string]struct{}, len(groups))
	for _, g := range groups {
		if _, dup := seen[g.Name]; dup {
			return fmt.Errorf("グループ名 '%s' が重複しています", g.Name)
		}
		seen[g.Name] = struct{}{}
	}
	return nil
}

type resolveState int

const (
	stateUnvisited resolveState = iota
	stateVisiting
	stateDone
)

// roomResolver は式を持つグループを依存関係順に解決します。
type roomResolver struct {
	groups []Group
	state  []resolveState
}

// resolveRoomExpressions は集合演算式を解決し、auto タスクの人数を算出します。
func resolveRoomExpressions(groups []Group) error {
	r := &roomResolver{
		groups: groups,
		state:  make([]resolveState, len(groups)),
	}
	// 式の解決（参照先が式を持つ場合は先にそちらを解決する）
	for i := range groups {
		if err := r.resolve(i); err != nil {
			return err
		}
	}

	return assignAutoTasks(groups)
}

// resolve は i 番目のグループの式を解決します。循環参照はエラーとします。
func (r *roomResolver) resolve(i int) error {
	switch r.state[i] {
	case stateDone:
		return nil
	case stateVisiting:
		return fmt.Errorf("グループ %s の式が循環参照しています", r.groups[i].Name)
	}

	r.state[i] = stateVisiting
	if expr := r.groups[i].rawRoomsExpr; expr != "" {
		rooms, err := r.evalExpr(expr)
		if err != nil {
			return fmt.Errorf("グループ %s の式 '%s': %w", r.groups[i].Name, expr, err)
		}
		r.groups[i].Rooms = rooms
	}
	r.state[i] = stateDone
	return nil
}

// evalExpr は "A - B - C" 形式の差集合式を評価します。
// 演算子 "-" は前後を空白で区切る必要があります（グループ名にハイフンを含められるようにするため）。
func (r *roomResolver) evalExpr(expr string) ([]uint, error) {
	tokens := strings.Fields(expr)
	if len(tokens) == 0 {
		return nil, errors.New("式が空です")
	}

	// 項と演算子が交互に並んでいることを検証
	var operands []string
	for i, tok := range tokens {
		isOp := tok == "-"
		if isOp != (i%2 == 1) {
			return nil, errors.New("式の形式が不正です ('名前 - 名前 - ...' のように空白区切りで指定してください)")
		}
		if !isOp {
			operands = append(operands, tok)
		}
	}
	if len(tokens)%2 == 0 {
		return nil, errors.New("式が演算子で終わっています")
	}

	baseRooms, err := r.collect(operands[0])
	if err != nil {
		return nil, err
	}

	removeSet := make(map[uint]struct{})
	for _, operand := range operands[1:] {
		removeRooms, err := r.collect(operand)
		if err != nil {
			return nil, err
		}
		for _, room := range removeRooms {
			removeSet[room] = struct{}{}
		}
	}

	// 元の baseRooms の順序を保持し、重複を除いて残った部屋を格納
	var result []uint
	seen := make(map[uint]struct{}, len(baseRooms))
	for _, room := range baseRooms {
		if _, removed := removeSet[room]; removed {
			continue
		}
		if _, dup := seen[room]; dup {
			continue
		}
		seen[room] = struct{}{}
		result = append(result, room)
	}
	return result, nil
}

// collect は identifier に該当するグループ（解決済み）の部屋を連結して返します。
// 該当するグループが存在しない場合はエラーとします。
func (r *roomResolver) collect(identifier string) ([]uint, error) {
	indices := findGroups(r.groups, identifier)
	if len(indices) == 0 {
		return nil, fmt.Errorf("参照先 '%s' が存在しません", identifier)
	}

	var rooms []uint
	for _, idx := range indices {
		if err := r.resolve(idx); err != nil {
			return nil, err
		}
		rooms = append(rooms, r.groups[idx].Rooms...)
	}
	return rooms, nil
}

// findGroups は identifier に該当するグループのインデックスを返します。
// グループ名の完全一致、および配下のグループ（例: "posts" -> "posts.leader", "posts.mentor" など）が対象です。
func findGroups(groups []Group, identifier string) []int {
	var result []int
	prefix := identifier + "."
	for i, g := range groups {
		if g.Name == identifier || strings.HasPrefix(g.Name, prefix) {
			result = append(result, i)
		}
	}
	return result
}

// assignAutoTasks は auto タスクの人数を算出します（部屋数 - 他の固定タスク合計）。
// 残りが auto タスクの数で割り切れない場合は、端数を定義順の先頭の auto タスクから 1 人ずつ配分します。
func assignAutoTasks(groups []Group) error {
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
		if len(autoTasks) == 0 {
			continue
		}

		roomCount := uint(len(groups[i].Rooms))
		if roomCount < fixedTotal {
			return fmt.Errorf("グループ %s: 固定タスクの合計(%d)が部屋数(%d)を超えているため auto を算出できません",
				groups[i].Name, fixedTotal, roomCount)
		}

		remaining := roomCount - fixedTotal
		n := uint(len(autoTasks))
		perTask, extra := remaining/n, remaining%n
		for k, at := range autoTasks {
			at.Number = perTask
			if uint(k) < extra {
				at.Number++
			}
		}
	}
	return nil
}

package config_test

import (
	"slices"
	"testing"

	"cleaning-table/internal/config"
	"cleaning-table/internal/domain"
	"cleaning-table/internal/validator"
)

func TestLoadFromFile_IdealTestYAML(t *testing.T) {
	rules, err := config.LoadFromFile("../../ideal_test.yaml")
	if err != nil {
		t.Fatalf("failed to load ideal_test.yaml: %v", err)
	}

	// posts 6ルール + common 9ルール = 計15ルール
	const expectedRules = 15
	if len(rules) != expectedRules {
		t.Fatalf("expected %d rules, got %d", expectedRules, len(rules))
	}

	// 全ルールで部屋数とタスク数が一致していること
	totalAssignedRooms := 0
	for i, r := range rules {
		if len(r.Rooms) == 0 {
			t.Errorf("rule[%d] has 0 rooms", i)
		}
		if len(r.Rooms) != len(r.Tasks) {
			t.Errorf("rule[%d] rooms count (%d) != tasks count (%d)", i, len(r.Rooms), len(r.Tasks))
		}
		totalAssignedRooms += len(r.Rooms)
	}

	// 1F common の部屋番号検証
	// all_rooms.1F (49) - facility (22) - posts (18) = [104, 105, 106, 107, 108, 111, 133, 134, 144] (9部屋)
	expected1FRooms := []int{104, 105, 106, 107, 108, 111, 133, 134, 144}
	var found1FRule bool
	for _, r := range rules {
		if slices.Equal(r.Rooms, expected1FRooms) {
			found1FRule = true
			// タスクの検証: フロア1, トイレ1期1, トイレ2期2, ゴミ分別2, 自室清掃(auto) 9 - 6 = 3
			taskCounts := make(map[string]int)
			for _, task := range r.Tasks {
				taskCounts[task]++
			}
			if taskCounts["フロア"] != 1 {
				t.Errorf("expected 1 'フロア', got %d", taskCounts["フロア"])
			}
			if taskCounts["トイレ1期"] != 1 {
				t.Errorf("expected 1 'トイレ1期', got %d", taskCounts["トイレ1期"])
			}
			if taskCounts["トイレ2期"] != 2 {
				t.Errorf("expected 2 'トイレ2期', got %d", taskCounts["トイレ2期"])
			}
			if taskCounts["ゴミ分別"] != 2 {
				t.Errorf("expected 2 'ゴミ分別', got %d", taskCounts["ゴミ分別"])
			}
			if taskCounts["自室清掃"] != 3 {
				t.Errorf("expected 3 '自室清掃' (auto), got %d", taskCounts["自室清掃"])
			}
			break
		}
	}
	if !found1FRule {
		t.Errorf("1F common rule with rooms %v not found", expected1FRooms)
	}

	// 既存バリデータ CheckDuplicatesAndSparse を用いた全室カバー率の検証
	// 全部屋ルール（49部屋×9階 = 441部屋）を先頭に追加して全室過不足なくカバーされているかを検証
	allRoomsSlice := make([]int, 0, 441)
	for floor := 1; floor <= 9; floor++ {
		for room := 1; room <= 49; room++ {
			allRoomsSlice = append(allRoomsSlice, floor*100+room)
		}
	}

	validationRules := append([]domain.AssignRule{
		{Rooms: allRoomsSlice, Tasks: make([]string, len(allRoomsSlice))},
		// facility ルール（33部屋）も部分ルールとして追加
		{Rooms: []int{101, 102, 109, 117, 127, 128, 129, 130, 131, 132, 135, 136, 137, 138, 139, 140, 141, 142, 145, 146, 147, 148, 202, 241, 242, 445, 541, 542, 646, 841, 842, 941, 942}, Tasks: make([]string, 33)},
	}, rules...)

	if err := validator.CheckDuplicatesAndSparse(validationRules); err != nil {
		t.Errorf("CheckDuplicatesAndSparse failed on ideal_test.yaml rules: %v", err)
	}
}

func TestLoadFromBytes_BasicAndAuto(t *testing.T) {
	yamlData := `
groups:
  floor1: [101, 102, 103, 104]
  exempt: [101]

assignments:
  main:
    rule1:
      rooms: "floor1 - exempt"
      tasks:
        掃除機: 1
        自室: auto
`
	rules, err := config.LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	expectedRooms := []int{102, 103, 104}
	if !slices.Equal(rules[0].Rooms, expectedRooms) {
		t.Errorf("rooms = %v, want %v", rules[0].Rooms, expectedRooms)
	}

	expectedTasks := []string{"掃除機", "自室", "自室"}
	if !slices.Equal(rules[0].Tasks, expectedTasks) {
		t.Errorf("tasks = %v, want %v", rules[0].Tasks, expectedTasks)
	}
}

func TestLoadFromBytes_ExactMatchWithoutAuto(t *testing.T) {
	yamlData := `
assignments:
  direct:
    rule:
      rooms: [101, 102]
      tasks:
        タスクA: 1
        タスクB: 1
`
	rules, err := config.LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if len(rules[0].Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(rules[0].Tasks))
	}
}

func TestLoadFromBytes_Errors(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "assignments セクションなし",
			yaml:    "groups:\n  a: [101]",
			wantErr: true,
		},
		{
			name: "auto が複数存在",
			yaml: `
assignments:
  r:
    rooms: [101, 102]
    tasks:
      T1: auto
      T2: auto
`,
			wantErr: true,
		},
		{
			name: "固定人数が部屋数を超過",
			yaml: `
assignments:
  r:
    rooms: [101]
    tasks:
      T1: 2
`,
			wantErr: true,
		},
		{
			name: "未定義のシンボルを参照する式",
			yaml: `
assignments:
  r:
    rooms: "unknown_group - other"
    tasks:
      T1: auto
`,
			wantErr: true,
		},
		{
			name: "負のタスク人数",
			yaml: `
assignments:
  r:
    rooms: [101]
    tasks:
      T1: -1
`,
			wantErr: true,
		},
		{
			name: "タスク人数と部屋数が不一致（autoなし）",
			yaml: `
assignments:
  r:
    rooms: [101, 102]
    tasks:
      T1: 1
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.LoadFromBytes([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package config_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	config "cleaning-table/internal"

	"gopkg.in/yaml.v3"
)

func TestConfigUnmarshalYAML(t *testing.T) {
	data, err := os.ReadFile("../testdata/config_v2.yaml")
	if err != nil {
		t.Fatalf("failed to read testdata/config_v2.yaml: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	if len(cfg.Groups) == 0 {
		t.Fatal("expected groups to be non-empty")
	}

	t.Logf("Total groups loaded: %d", len(cfg.Groups))

	// グループの検証
	groupMap := make(map[string]config.Group)
	for _, g := range cfg.Groups {
		groupMap[g.Name] = g
		t.Logf("Group: %-20s Rooms: %2d, Tasks: %d", g.Name, len(g.Rooms), len(g.Tasks))
	}

	// 1. facility グループの検証
	facility, ok := groupMap["facility"]
	if !ok {
		t.Error("facility group not found")
	} else if len(facility.Rooms) == 0 {
		t.Error("facility rooms should not be empty")
	}

	// 2. leader グループの検証
	leader, ok := groupMap["posts.leader"]
	if !ok {
		t.Error("posts.leader group not found")
	} else {
		if len(leader.Rooms) != 1 || leader.Rooms[0] != 724 {
			t.Errorf("expected leader room 724, got %v", leader.Rooms)
		}
		if len(leader.Tasks) != 1 || leader.Tasks[0].Name != "寮長" || leader.Tasks[0].Number != 1 {
			t.Errorf("expected task 寮長: 1, got %v", leader.Tasks)
		}
	}

	// 3. common.1F グループの検証（集合演算の式と auto タスクの計算確認）
	common1F, ok := groupMap["common.1F"]
	if !ok {
		t.Error("common.1F group not found")
	} else {
		t.Logf("common.1F rooms count: %d, tasks: %+v", len(common1F.Rooms), common1F.Tasks)
		if len(common1F.Rooms) == 0 {
			t.Error("common.1F rooms should be computed from expression")
		}

		// 1F は部屋14室、固定タスク14人のため auto は 0人
		foundAuto := false
		for _, task := range common1F.Tasks {
			if task.Name == "自室清掃" {
				foundAuto = true
				if !task.IsAuto {
					t.Error("expected 自室清掃 to be IsAuto = true")
				}
				if task.Number != 0 {
					t.Errorf("expected 自室清掃 in 1F to have Number 0, got %d", task.Number)
				}
			}
		}
		if !foundAuto {
			t.Error("task 自室清掃 not found in common.1F")
		}
	}

	// 4. common.2F グループの検証（34室 - 固定23人 = 自室清掃11人）
	common2F, ok := groupMap["common.2F"]
	if !ok {
		t.Error("common.2F group not found")
	} else {
		for _, task := range common2F.Tasks {
			if task.Name == "自室清掃" {
				if task.Number != 11 {
					t.Errorf("expected 自室清掃 in 2F to have Number 11, got %d", task.Number)
				}
			}
		}
	}
}

func unmarshal(t *testing.T, src string) (map[string]config.Group, error) {
	t.Helper()

	var cfg config.Config
	if err := yaml.Unmarshal([]byte(src), &cfg); err != nil {
		return nil, err
	}
	m := make(map[string]config.Group, len(cfg.Groups))
	for _, g := range cfg.Groups {
		m[g.Name] = g
	}
	return m, nil
}

func TestConfigUnmarshalYAML_Errors(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
	}{
		{"引き算側の参照先が存在しない", "a: [1,2]\nc:\n  rooms: \"a - bb\"\n", "'bb'"},
		{"循環参照", "a:\n  rooms: \"b\"\nb:\n  rooms: \"a\"\n", "循環参照"},
		{"自己参照", "c:\n  x:\n    rooms: \"c\"\n", "循環参照"},
		{"固定タスクが部屋数超過", "a: [1,2]\nc:\n  rooms: \"a\"\n  tasks:\n    t1: 5\n    t2: auto\n", "超えている"},
		{"タスク値が null", "a:\n  rooms: [1]\n  tasks:\n    t1:\n", "値が空"},
		{"rooms が null", "a:\n  rooms:\n  tasks:\n    t: auto\n", "rooms が空"},
		{"rooms が数値", "a:\n  rooms: 101\n", "配列または式"},
		{"未知のキー", "a:\n  room: [1]\n  tasks:\n    t: auto\n", "未知のキー 'room'"},
		{"グループが null", "a:\n", "定義が空"},
		{"トップレベルのキー重複", "a: [1]\na: [2]\n", "重複"},
		{"ドット付きキーとネストの衝突", "a.b: [1]\na:\n  b: [2]\n", "グループ名 'a.b' が重複"},
		{"タスク名の重複", "a:\n  rooms: [1]\n  tasks:\n    t: 1\n    t: 2\n", "重複"},
		{"部屋番号の重複", "a: [1, 2, 1]\n", "部屋番号 1 が重複"},
		{"式が演算子で始まる", "a: [1]\nc:\n  rooms: \"- a\"\n", "形式が不正"},
		{"式が演算子で終わる", "a: [1]\nc:\n  rooms: \"a -\"\n", "演算子で終わって"},
		{"演算子の連続", "a: [1]\nc:\n  rooms: \"a - - a\"\n", "形式が不正"},
		{"マージキー", "base: &b\n  rooms: [1]\nx:\n  <<: *b\n", "マージキー"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := unmarshal(t, tt.src)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestConfigUnmarshalYAML_Expressions(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		group     string
		wantRooms []uint
	}{
		{
			"後方で定義された式グループの参照は依存順に解決される",
			"a: [1,2,3,4]\nb: [1]\nc:\n  rooms: \"a - d\"\nd:\n  rooms: \"b\"\n",
			"c",
			[]uint{2, 3, 4},
		},
		{
			"前方一致で集めた部屋の重複は除去される",
			"p:\n  x: [1,2]\n  y: [2,3]\nq: [9]\nc:\n  rooms: \"p - q\"\n",
			"c",
			[]uint{1, 2, 3},
		},
		{
			"完全一致と配下グループの両方が対象になる",
			"p: [1]\np.x: [2]\nc:\n  rooms: \"p\"\n",
			"c",
			[]uint{1, 2},
		},
		{
			"ハイフンを含むグループ名を参照できる",
			"all-rooms: [1,2,3]\nx: [1]\nc:\n  rooms: \"all-rooms - x\"\n",
			"c",
			[]uint{2, 3},
		},
		{
			"YAML アンカー/エイリアスを解決できる",
			"a: &x [1,2]\nb: *x\n",
			"b",
			[]uint{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, err := unmarshal(t, tt.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := groups[tt.group].Rooms; !reflect.DeepEqual(got, tt.wantRooms) {
				t.Errorf("rooms of %s: want %v, got %v", tt.group, tt.wantRooms, got)
			}
		})
	}
}

func TestConfigUnmarshalYAML_AutoTasks(t *testing.T) {
	// 残り 5 室を auto 2 つで分ける → 端数は先頭の auto に配分 (3, 2)
	groups, err := unmarshal(t, "a: [1,2,3,4,5,6,7]\nc:\n  rooms: \"a\"\n  tasks:\n    t0: 2\n    t1: auto\n    t2: auto\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tasks := groups["c"].Tasks
	var total uint
	for _, task := range tasks {
		total += task.Number
	}
	if total != 7 {
		t.Errorf("all rooms should be assigned: total=%d tasks=%+v", total, tasks)
	}
	if tasks[1].Number != 3 || tasks[2].Number != 2 {
		t.Errorf("unexpected distribution: %+v", tasks)
	}
}

package config_test

import (
	"os"
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

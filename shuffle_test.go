package main

import (
	"fmt"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
)

func TestShuffleTask_Basic(t *testing.T) {
	rules := []assignRule{
		{
			rooms: mapset.NewSet(101, 102, 103),
			tasks: []string{"ゴミ出し", "掃除機", "風呂掃除"},
		},
	}

	history := make(assignHistory)
	results, newHistory, err := shuffleTask(rules, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// 各部屋にタスクが1つずつ割り当てられていること
	assignedRooms := make(map[int]bool)
	assignedTasks := make(map[string]bool)
	for _, res := range results {
		if assignedRooms[res.room] {
			t.Errorf("duplicate room assigned: %d", res.room)
		}
		if assignedTasks[res.task] {
			t.Errorf("duplicate task assigned: %s", res.task)
		}
		assignedRooms[res.room] = true
		assignedTasks[res.task] = true

		if newHistory[res.room][res.task] != 1 {
			t.Errorf("expected count 1 for room %d and task %s, got %d", res.room, res.task, newHistory[res.room][res.task])
		}
	}
}

func TestShuffleTask_DuplicateTasks(t *testing.T) {
	// 同じタスク（例: ゴミ分別が2人分）が含まれるケース
	rules := []assignRule{
		{
			rooms: mapset.NewSet(101, 102, 103),
			tasks: []string{"ゴミ分別", "シャワー室", "ゴミ分別"},
		},
	}

	history := make(assignHistory)
	results, newHistory, err := shuffleTask(rules, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	taskCounts := make(map[string]int)
	roomAssigned := make(map[int]bool)
	for _, res := range results {
		if roomAssigned[res.room] {
			t.Errorf("duplicate room assigned: %d", res.room)
		}
		roomAssigned[res.room] = true
		taskCounts[res.task]++
	}

	if taskCounts["ゴミ分別"] != 2 {
		t.Errorf("expected 2 'ゴミ分別', got %d", taskCounts["ゴミ分別"])
	}
	if taskCounts["シャワー室"] != 1 {
		t.Errorf("expected 1 'シャワー室', got %d", taskCounts["シャワー室"])
	}

	// 履歴の更新も確認
	gomiTotal := 0
	for room := range roomAssigned {
		gomiTotal += newHistory[room]["ゴミ分別"]
	}
	if gomiTotal != 2 {
		t.Errorf("expected total 2 in history for 'ゴミ分別', got %d", gomiTotal)
	}
}

func TestShuffleTask_ErrorMismatch(t *testing.T) {
	rules := []assignRule{
		{
			rooms: mapset.NewSet(101, 102),
			tasks: []string{"ゴミ出し", "掃除機", "風呂掃除"},
		},
	}

	_, _, err := shuffleTask(rules, nil)
	if err == nil {
		t.Fatal("expected error for mismatched rooms and tasks, got nil")
	}
}

func TestShuffleTask_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		rules       []assignRule
		history     assignHistory
		wantErr     bool
		wantResults int
	}{
		{
			name:        "rules が nil",
			rules:       nil,
			history:     nil,
			wantErr:     false,
			wantResults: 0,
		},
		{
			name:        "rules が空スライス",
			rules:       []assignRule{},
			history:     nil,
			wantErr:     false,
			wantResults: 0,
		},
		{
			name: "rule.rooms が nil",
			rules: []assignRule{
				{rooms: nil, tasks: []string{"A"}},
			},
			history: nil,
			wantErr: true,
		},
		{
			name: "rule.tasks が nil",
			rules: []assignRule{
				{rooms: mapset.NewSet(1), tasks: nil},
			},
			history: nil,
			wantErr: true,
		},
		{
			name: "部屋とタスクが両方空（要素数0）",
			rules: []assignRule{
				{rooms: mapset.NewSet[int](), tasks: []string{}},
			},
			history: nil,
			wantErr: true,
		},
		{
			name: "history が nil",
			rules: []assignRule{
				{rooms: mapset.NewSet(1), tasks: []string{"A"}},
			},
			history:     nil,
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "history[room] が nil",
			rules: []assignRule{
				{rooms: mapset.NewSet(1), tasks: []string{"A"}},
			},
			history:     assignHistory{1: nil},
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "history に負の値が含まれる",
			rules: []assignRule{
				{rooms: mapset.NewSet(1), tasks: []string{"A"}},
			},
			history:     assignHistory{1: {"A": -10}},
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "history に非常に大きな値が含まれる",
			rules: []assignRule{
				{rooms: mapset.NewSet(1), tasks: []string{"A"}},
			},
			history:     assignHistory{1: {"A": 1000000}},
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "タスク名が空文字列や特殊文字",
			rules: []assignRule{
				{rooms: mapset.NewSet(1, 2), tasks: []string{"", "🧹✨\n\t"}},
			},
			history:     nil,
			wantErr:     false,
			wantResults: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()

			results, newH, err := shuffleTask(tt.rules, tt.history)
			if (err != nil) != tt.wantErr {
				t.Fatalf("shuffleTask() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(results) != tt.wantResults {
					t.Errorf("expected %d results, got %d", tt.wantResults, len(results))
				}
				if newH == nil {
					t.Error("expected non-nil newHistory")
				}
			}
		})
	}
}

func TestShuffleTask_FairnessSimulation(t *testing.T) {
	// 4部屋・4タスクで 40 回割り当てを繰り返す
	rooms := []int{1, 2, 3, 4}
	tasks := []string{"TaskA", "TaskB", "TaskC", "TaskD"}

	rules := []assignRule{
		{
			rooms: mapset.NewSet(rooms...),
			tasks: tasks,
		},
	}

	history := make(assignHistory)
	const rounds = 40

	for range rounds {
		results, newHistory, err := shuffleTask(rules, history)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != len(rooms) {
			t.Fatalf("expected %d results, got %d", len(rooms), len(results))
		}
		history = newHistory
	}

	// 40回実行後、各部屋の各タスクは 40 / 4 = 10 回ずつ行われているはず
	for _, room := range rooms {
		for _, task := range tasks {
			count := history[room][task]
			if count != rounds/len(tasks) {
				t.Errorf("room %d task %s count = %d, expected exactly %d", room, task, count, rounds/len(tasks))
			}
		}
	}
}

func TestShuffleTask_PrioritizeInfrequentTask(t *testing.T) {
	// 部屋1はすでに TaskA を 10 回、TaskB を 0 回行っている
	// 部屋2はすでに TaskA を 0 回、TaskB を 10 回行っている
	rules := []assignRule{
		{
			rooms: mapset.NewSet(1, 2),
			tasks: []string{"TaskA", "TaskB"},
		},
	}

	history := assignHistory{
		1: {"TaskA": 10, "TaskB": 0},
		2: {"TaskA": 0, "TaskB": 10},
	}

	results, _, err := shuffleTask(rules, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, res := range results {
		if res.room == 1 && res.task != "TaskB" {
			t.Errorf("room 1 should be assigned TaskB, got %s", res.task)
		}
		if res.room == 2 && res.task != "TaskA" {
			t.Errorf("room 2 should be assigned TaskA, got %s", res.task)
		}
	}
}

func TestShuffleTask_Immutability(t *testing.T) {
	rules := []assignRule{
		{
			rooms: mapset.NewSet(1),
			tasks: []string{"TaskA"},
		},
	}

	origHistory := assignHistory{
		1: {"TaskA": 1},
	}

	_, newHistory, err := shuffleTask(rules, origHistory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if origHistory[1]["TaskA"] != 1 {
		t.Errorf("origHistory was mutated: got %d", origHistory[1]["TaskA"])
	}
	if newHistory[1]["TaskA"] != 2 {
		t.Errorf("newHistory not updated: got %d", newHistory[1]["TaskA"])
	}
}

func TestVerifyFairness(t *testing.T) {
	t.Run("長期均等シミュレーション (10部屋×10タスク, 100回実行)", func(t *testing.T) {
		n := 10
		rounds := 100
		roomSlice := make([]int, n)
		taskSlice := make([]string, n)
		for i := range n {
			roomSlice[i] = 101 + i
			taskSlice[i] = fmt.Sprintf("Task_%c", 'A'+i)
		}

		rules := []assignRule{
			{
				rooms: mapset.NewSet(roomSlice...),
				tasks: taskSlice,
			},
		}

		history := make(assignHistory)
		for range rounds {
			_, newH, err := shuffleTask(rules, history)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			history = newH
		}

		for _, room := range roomSlice {
			for _, task := range taskSlice {
				c := history[room][task]
				if c != rounds/n {
					t.Errorf("Room %d Task %s count = %d, expected %d", room, task, c, rounds/n)
				}
			}
		}
	})

	t.Run("初期偏りからの補正シミュレーション (4部屋×4タスク, 偏り10回から20回追加)", func(t *testing.T) {
		roomSlice := []int{1, 2, 3, 4}
		taskSlice := []string{"Task_A", "Task_B", "Task_C", "Task_D"}
		rules := []assignRule{
			{
				rooms: mapset.NewSet(roomSlice...),
				tasks: taskSlice,
			},
		}

		history := assignHistory{
			1: {"Task_A": 10, "Task_B": 0, "Task_C": 0, "Task_D": 0},
			2: {"Task_A": 0, "Task_B": 10, "Task_C": 0, "Task_D": 0},
			3: {"Task_A": 0, "Task_B": 0, "Task_C": 10, "Task_D": 0},
			4: {"Task_A": 0, "Task_B": 0, "Task_C": 0, "Task_D": 10},
		}

		for range 20 {
			_, newH, err := shuffleTask(rules, history)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			history = newH
		}

		// 20回追加後、各部屋で過去0回だった3タスクはおよそ6〜7回行われ、10回だったタスクは0回（または最小限）追加される
		for _, room := range roomSlice {
			if room == 1 {
				if history[1]["Task_A"] > 11 {
					t.Errorf("Room 1 Task_A was already 10, should not be frequently assigned, got %d", history[1]["Task_A"])
				}
				if history[1]["Task_B"] < 6 || history[1]["Task_C"] < 6 || history[1]["Task_D"] < 6 {
					t.Errorf("Room 1 other tasks should be prioritized, got B=%d, C=%d, D=%d", history[1]["Task_B"], history[1]["Task_C"], history[1]["Task_D"])
				}
			}
		}
	})

	t.Run("端数回数での差が最大1回であることの検証 (4部屋×4タスク, 15回実行)", func(t *testing.T) {
		roomSlice := []int{1, 2, 3, 4}
		taskSlice := []string{"Task_A", "Task_B", "Task_C", "Task_D"}
		rules := []assignRule{
			{
				rooms: mapset.NewSet(roomSlice...),
				tasks: taskSlice,
			},
		}

		history := make(assignHistory)
		for range 15 {
			_, newH, err := shuffleTask(rules, history)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			history = newH
		}

		for _, room := range roomSlice {
			for _, task := range taskSlice {
				c := history[room][task]
				if c != 3 && c != 4 {
					t.Errorf("Room %d Task %s count = %d, expected 3 or 4", room, task, c)
				}
			}
		}
	})
}

func BenchmarkShuffleTask(b *testing.B) {
	// 20ルール、各ルール50部屋のベンチマーク
	const numRules = 20
	const numRooms = 50

	rules := make([]assignRule, numRules)
	for i := range numRules {
		roomSlice := make([]int, numRooms)
		taskSlice := make([]string, numRooms)
		for j := range numRooms {
			roomSlice[j] = i*1000 + j
			taskSlice[j] = fmt.Sprintf("task_%d_%d", i, j)
		}
		rules[i] = assignRule{
			rooms: mapset.NewSet(roomSlice...),
			tasks: taskSlice,
		}
	}

	history := make(assignHistory)

	b.ResetTimer()
	for b.Loop() {
		_, _, err := shuffleTask(rules, history)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

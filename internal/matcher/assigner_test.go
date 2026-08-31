package matcher_test

import (
	"fmt"
	"testing"

	"cleaning-table/internal/domain"
	"cleaning-table/internal/matcher"
)

func TestShuffleTask_Basic(t *testing.T) {
	rules := []domain.AssignRule{
		{
			Rooms: []int{101, 102, 103},
			Tasks: []string{"ゴミ出し", "掃除機", "風呂掃除"},
		},
	}

	history := make(domain.AssignHistory)
	results, newHistory, err := matcher.ShuffleTask(rules, history)
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
		if assignedRooms[res.Room] {
			t.Errorf("duplicate room assigned: %d", res.Room)
		}
		if assignedTasks[res.Task] {
			t.Errorf("duplicate task assigned: %s", res.Task)
		}
		assignedRooms[res.Room] = true
		assignedTasks[res.Task] = true

		if newHistory[res.Room][res.Task] != 1 {
			t.Errorf("expected count 1 for room %d and task %s, got %d", res.Room, res.Task, newHistory[res.Room][res.Task])
		}
	}
}

func TestShuffleTask_DuplicateTasks(t *testing.T) {
	// 同じタスク（例: ゴミ分別が2人分）が含まれるケース
	rules := []domain.AssignRule{
		{
			Rooms: []int{101, 102, 103},
			Tasks: []string{"ゴミ分別", "シャワー室", "ゴミ分別"},
		},
	}

	history := make(domain.AssignHistory)
	results, newHistory, err := matcher.ShuffleTask(rules, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	taskCounts := make(map[string]int)
	roomAssigned := make(map[int]bool)
	for _, res := range results {
		if roomAssigned[res.Room] {
			t.Errorf("duplicate room assigned: %d", res.Room)
		}
		roomAssigned[res.Room] = true
		taskCounts[res.Task]++
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

func TestShuffleTask_DuplicateRoomInRule(t *testing.T) {
	// 同一ルール内に部屋番号の重複がある場合
	rules := []domain.AssignRule{
		{
			Rooms: []int{101, 101, 102},
			Tasks: []string{"A", "B", "C"},
		},
	}

	_, _, err := matcher.ShuffleTask(rules, nil)
	if err == nil {
		t.Fatal("expected error for duplicate room in rule, got nil")
	}
}

func TestShuffleTask_OrderInvariance(t *testing.T) {
	// 部屋番号の指定順序が昇順でも降順でも公平に割り当てられること
	rulesAsc := []domain.AssignRule{
		{
			Rooms: []int{101, 102, 103},
			Tasks: []string{"A", "B", "C"},
		},
	}
	rulesDesc := []domain.AssignRule{
		{
			Rooms: []int{103, 102, 101},
			Tasks: []string{"A", "B", "C"},
		},
	}

	historyAsc := make(domain.AssignHistory)
	historyDesc := make(domain.AssignHistory)

	const rounds = 30
	for range rounds {
		_, newHAsc, err := matcher.ShuffleTask(rulesAsc, historyAsc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		historyAsc = newHAsc

		_, newHDesc, err := matcher.ShuffleTask(rulesDesc, historyDesc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		historyDesc = newHDesc
	}

	for _, room := range []int{101, 102, 103} {
		for _, task := range []string{"A", "B", "C"} {
			if historyAsc[room][task] != 10 {
				t.Errorf("Asc: room %d task %s count = %d, expected 10", room, task, historyAsc[room][task])
			}
			if historyDesc[room][task] != 10 {
				t.Errorf("Desc: room %d task %s count = %d, expected 10", room, task, historyDesc[room][task])
			}
		}
	}
}

func TestShuffleTask_ErrorMismatch(t *testing.T) {
	rules := []domain.AssignRule{
		{
			Rooms: []int{101, 102},
			Tasks: []string{"ゴミ出し", "掃除機", "風呂掃除"},
		},
	}

	_, _, err := matcher.ShuffleTask(rules, nil)
	if err == nil {
		t.Fatal("expected error for mismatched rooms and tasks, got nil")
	}
}

func TestShuffleTask_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		rules       []domain.AssignRule
		history     domain.AssignHistory
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
			rules:       []domain.AssignRule{},
			history:     nil,
			wantErr:     false,
			wantResults: 0,
		},
		{
			name: "rule.Rooms が nil",
			rules: []domain.AssignRule{
				{Rooms: nil, Tasks: []string{"A"}},
			},
			history: nil,
			wantErr: true,
		},
		{
			name: "rule.Tasks が nil",
			rules: []domain.AssignRule{
				{Rooms: []int{1}, Tasks: nil},
			},
			history: nil,
			wantErr: true,
		},
		{
			name: "部屋とタスクが両方空（要素数0）",
			rules: []domain.AssignRule{
				{Rooms: []int{}, Tasks: []string{}},
			},
			history: nil,
			wantErr: true,
		},
		{
			name: "history が nil",
			rules: []domain.AssignRule{
				{Rooms: []int{1}, Tasks: []string{"A"}},
			},
			history:     nil,
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "history[room] が nil",
			rules: []domain.AssignRule{
				{Rooms: []int{1}, Tasks: []string{"A"}},
			},
			history:     domain.AssignHistory{1: nil},
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "history に負の値が含まれる",
			rules: []domain.AssignRule{
				{Rooms: []int{1}, Tasks: []string{"A"}},
			},
			history:     domain.AssignHistory{1: {"A": -10}},
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "history に非常に大きな値が含まれる",
			rules: []domain.AssignRule{
				{Rooms: []int{1}, Tasks: []string{"A"}},
			},
			history:     domain.AssignHistory{1: {"A": 1000000}},
			wantErr:     false,
			wantResults: 1,
		},
		{
			name: "タスク名が空文字列や特殊文字",
			rules: []domain.AssignRule{
				{Rooms: []int{1, 2}, Tasks: []string{"", "🧹✨\n\t"}},
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

			results, newH, err := matcher.ShuffleTask(tt.rules, tt.history)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ShuffleTask() error = %v, wantErr %v", err, tt.wantErr)
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

	rules := []domain.AssignRule{
		{
			Rooms: rooms,
			Tasks: tasks,
		},
	}

	history := make(domain.AssignHistory)
	const rounds = 40

	for range rounds {
		results, newHistory, err := matcher.ShuffleTask(rules, history)
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
	rules := []domain.AssignRule{
		{
			Rooms: []int{1, 2},
			Tasks: []string{"TaskA", "TaskB"},
		},
	}

	history := domain.AssignHistory{
		1: {"TaskA": 10, "TaskB": 0},
		2: {"TaskA": 0, "TaskB": 10},
	}

	results, _, err := matcher.ShuffleTask(rules, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, res := range results {
		if res.Room == 1 && res.Task != "TaskB" {
			t.Errorf("room 1 should be assigned TaskB, got %s", res.Task)
		}
		if res.Room == 2 && res.Task != "TaskA" {
			t.Errorf("room 2 should be assigned TaskA, got %s", res.Task)
		}
	}
}

func TestShuffleTask_Immutability(t *testing.T) {
	rules := []domain.AssignRule{
		{
			Rooms: []int{1},
			Tasks: []string{"TaskA"},
		},
	}

	origHistory := domain.AssignHistory{
		1: {"TaskA": 1},
	}

	_, newHistory, err := matcher.ShuffleTask(rules, origHistory)
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

		rules := []domain.AssignRule{
			{
				Rooms: roomSlice,
				Tasks: taskSlice,
			},
		}

		history := make(domain.AssignHistory)
		for range rounds {
			_, newH, err := matcher.ShuffleTask(rules, history)
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
		rules := []domain.AssignRule{
			{
				Rooms: roomSlice,
				Tasks: taskSlice,
			},
		}

		history := domain.AssignHistory{
			1: {"Task_A": 10, "Task_B": 0, "Task_C": 0, "Task_D": 0},
			2: {"Task_A": 0, "Task_B": 10, "Task_C": 0, "Task_D": 0},
			3: {"Task_A": 0, "Task_B": 0, "Task_C": 10, "Task_D": 0},
			4: {"Task_A": 0, "Task_B": 0, "Task_C": 0, "Task_D": 10},
		}

		for range 20 {
			_, newH, err := matcher.ShuffleTask(rules, history)
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
		rules := []domain.AssignRule{
			{
				Rooms: roomSlice,
				Tasks: taskSlice,
			},
		}

		history := make(domain.AssignHistory)
		for range 15 {
			_, newH, err := matcher.ShuffleTask(rules, history)
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

	rules := make([]domain.AssignRule, numRules)
	for i := range numRules {
		roomSlice := make([]int, numRooms)
		taskSlice := make([]string, numRooms)
		for j := range numRooms {
			roomSlice[j] = i*1000 + j
			taskSlice[j] = fmt.Sprintf("task_%d_%d", i, j)
		}
		rules[i] = domain.AssignRule{
			Rooms: roomSlice,
			Tasks: taskSlice,
		}
	}

	history := make(domain.AssignHistory)

	b.ResetTimer()
	for b.Loop() {
		_, _, err := matcher.ShuffleTask(rules, history)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

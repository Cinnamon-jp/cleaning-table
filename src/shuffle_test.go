package src

import (
	"testing"
)

func TestAssignTasks(t *testing.T) {
	data := UnfoldedExcelData{
		RoomNumbers: [][]int{
			{101, 102, 103},
			{201, 202},
		},
		Tasks: [][]string{
			{"TaskA", "TaskB", "TaskC"},
			{"TaskD", "TaskE"},
		},
	}

	result := AssignTasks(data)

	// 列数の確認
	if len(result) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result))
	}

	// 各列の要素数の確認
	if len(result[0]) != 3 {
		t.Errorf("expected 3 assignments in col 0, got %d", len(result[0]))
	}
	if len(result[1]) != 2 {
		t.Errorf("expected 2 assignments in col 1, got %d", len(result[1]))
	}

	// 割り振られたタスクの集合が元のタスクの集合と一致するか確認する
	// column 0
	taskMap0 := make(map[string]int)
	for _, a := range result[0] {
		taskMap0[a.Task]++
	}
	for _, task := range data.Tasks[0] {
		if taskMap0[task] == 0 {
			t.Errorf("task %s is missing or insufficient in col 0", task)
		} else {
			taskMap0[task]--
		}
	}

	// column 1
	taskMap1 := make(map[string]int)
	for _, a := range result[1] {
		taskMap1[a.Task]++
	}
	for _, task := range data.Tasks[1] {
		if taskMap1[task] == 0 {
			t.Errorf("task %s is missing or insufficient in col 1", task)
		} else {
			taskMap1[task]--
		}
	}

	// 部屋番号がすべて含まれているか確認する
	// column 0
	roomMap0 := make(map[int]bool)
	for _, a := range result[0] {
		roomMap0[a.Room] = true
	}
	for _, room := range data.RoomNumbers[0] {
		if !roomMap0[room] {
			t.Errorf("room %d is missing in col 0", room)
		}
	}

	// column 1
	roomMap1 := make(map[int]bool)
	for _, a := range result[1] {
		roomMap1[a.Room] = true
	}
	for _, room := range data.RoomNumbers[1] {
		if !roomMap1[room] {
			t.Errorf("room %d is missing in col 1", room)
		}
	}
}

func TestGroupByFloor(t *testing.T) {
	t.Run("基本動作: 複数階にまたがるAssignmentを正しく階ごとに振り分ける", func(t *testing.T) {
		input := [][]Assignment{
			{
				{Room: 101, Task: "フロア"},
				{Room: 201, Task: "トイレ"},
			},
			{
				{Room: 102, Task: "ゴミ分別"},
				{Room: 203, Task: "洗濯室"},
			},
		}

		result := GroupByFloor(input)

		// 9階分のスライスが返ること
		if len(result) != 9 {
			t.Fatalf("expected 9 floors, got %d", len(result))
		}

		// 各階が49要素であること
		for i, floor := range result {
			if len(floor) != 49 {
				t.Errorf("floor %d: expected 49 rooms, got %d", i+1, len(floor))
			}
		}

		// 1Fの101号室にフロアが割り当てられていること
		if result[0][0].Task != "フロア" {
			t.Errorf("1F room 01: expected Task 'フロア', got '%s'", result[0][0].Task)
		}
		// 1Fの102号室にゴミ分別が割り当てられていること
		if result[0][1].Task != "ゴミ分別" {
			t.Errorf("1F room 02: expected Task 'ゴミ分別', got '%s'", result[0][1].Task)
		}
		// 2Fの201号室にトイレが割り当てられていること
		if result[1][0].Task != "トイレ" {
			t.Errorf("2F room 01: expected Task 'トイレ', got '%s'", result[1][0].Task)
		}
		// 2Fの203号室に洗濯室が割り当てられていること
		if result[1][2].Task != "洗濯室" {
			t.Errorf("2F room 03: expected Task '洗濯室', got '%s'", result[1][2].Task)
		}
	})

	t.Run("号室昇順ソート: 同一階内で号室番号の昇順にソートされていること", func(t *testing.T) {
		input := [][]Assignment{
			{
				{Room: 149, Task: "タスクA"},
				{Room: 101, Task: "タスクB"},
				{Room: 125, Task: "タスクC"},
			},
		}

		result := GroupByFloor(input)

		// 1F の各要素が号室昇順であること
		for i := 0; i < len(result[0])-1; i++ {
			if result[0][i].Room >= result[0][i+1].Room {
				t.Errorf("1F: room order violation at index %d: %d >= %d",
					i, result[0][i].Room, result[0][i+1].Room)
			}
		}

		// 各タスクが正しい号室に入っていること
		if result[0][0].Task != "タスクB" {
			t.Errorf("1F room 01: expected 'タスクB', got '%s'", result[0][0].Task)
		}
		if result[0][24].Task != "タスクC" {
			t.Errorf("1F room 25: expected 'タスクC', got '%s'", result[0][24].Task)
		}
		if result[0][48].Task != "タスクA" {
			t.Errorf("1F room 49: expected 'タスクA', got '%s'", result[0][48].Task)
		}
	})

	t.Run("空入力: 空のスライスでも9階分の空Taskスライスが返ること", func(t *testing.T) {
		input := [][]Assignment{}

		result := GroupByFloor(input)

		if len(result) != 9 {
			t.Fatalf("expected 9 floors, got %d", len(result))
		}

		for i, floor := range result {
			if len(floor) != 49 {
				t.Errorf("floor %d: expected 49 rooms, got %d", i+1, len(floor))
			}
			for _, a := range floor {
				if a.Task != "" {
					t.Errorf("floor %d room %d: expected empty task, got '%s'",
						i+1, a.Room%100, a.Task)
				}
			}
		}
	})

	t.Run("単一階: 全部屋が同一階に属する場合", func(t *testing.T) {
		input := [][]Assignment{
			{
				{Room: 501, Task: "タスクX"},
				{Room: 510, Task: "タスクY"},
				{Room: 549, Task: "タスクZ"},
			},
		}

		result := GroupByFloor(input)

		// 5F（インデックス4）にのみタスクが存在すること
		if result[4][0].Task != "タスクX" {
			t.Errorf("5F room 01: expected 'タスクX', got '%s'", result[4][0].Task)
		}
		if result[4][9].Task != "タスクY" {
			t.Errorf("5F room 10: expected 'タスクY', got '%s'", result[4][9].Task)
		}
		if result[4][48].Task != "タスクZ" {
			t.Errorf("5F room 49: expected 'タスクZ', got '%s'", result[4][48].Task)
		}

		// 他の階のタスクがすべて空文字列であること
		for i, floor := range result {
			if i == 4 {
				continue
			}
			for _, a := range floor {
				if a.Task != "" {
					t.Errorf("floor %d room %d: expected empty task, got '%s'",
						i+1, a.Room%100, a.Task)
				}
			}
		}
	})

	t.Run("全9階: 1F〜9Fすべてにデータがある場合", func(t *testing.T) {
		var input [][]Assignment
		var col []Assignment
		for floor := 1; floor <= 9; floor++ {
			col = append(col, Assignment{
				Room: floor*100 + 1,
				Task: "タスク",
			})
		}
		input = append(input, col)

		result := GroupByFloor(input)

		if len(result) != 9 {
			t.Fatalf("expected 9 floors, got %d", len(result))
		}

		// 各階の01号室にタスクが割り当てられていること
		for i := 0; i < 9; i++ {
			if result[i][0].Task != "タスク" {
				t.Errorf("floor %d room 01: expected 'タスク', got '%s'",
					i+1, result[i][0].Task)
			}
			// 02号室以降は空文字列であること
			for j := 1; j < 49; j++ {
				if result[i][j].Task != "" {
					t.Errorf("floor %d room %02d: expected empty task, got '%s'",
						i+1, j+1, result[i][j].Task)
				}
			}
		}
	})
}

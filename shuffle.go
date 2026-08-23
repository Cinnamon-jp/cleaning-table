package main

import (
	"fmt"
	"math/rand/v2"
)

func shuffleTask(
	rules []assignRule,
	history assignHistory,
) (result []assignResult, newHistory assignHistory, err error) {

	for i, rule := range rules {
		nRooms := rule.rooms.Cardinality()
		nTasks := rule.tasks.Cardinality()
		if nRooms != nTasks {
			return nil, nil, fmt.Errorf("number of rooms and tasks must match [%d != %d] (index %d)", nRooms, nTasks, i)
		}

		rooms := rule.rooms.ToSlice()
		tasks := rule.tasks.ToSlice()

		// tasksをシャッフル
		rand.Shuffle(len(tasks), func(i, j int) {
			tasks[i], tasks[j] = tasks[j], tasks[i]
		})

		for _, room := range rooms {
			if prevTasks := history[room]; prevTasks != nil {
			}
		}
	}
	return nil, nil, nil
}

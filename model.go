package main

import (
	mapset "github.com/deckarep/golang-set/v2"
)

type assignRule struct {
	rooms mapset.Set[int]
	tasks []string
}

type assignResult struct {
	room int
	task string
}

// room: (task: times)
type assignHistory map[int](map[string]int)

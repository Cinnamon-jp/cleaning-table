package config

type Config struct {
	Groups []Group
}

type Group struct {
	Name  string
	Rooms []uint
	Tasks []Task
}

type Task struct {
	Name   string
	Number uint
}

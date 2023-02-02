package model

type Task struct {
	Task    interface{} //Steps
	Desc    string
	Name    string
	Public  bool
	Ref     string
	RefDir  string
	Finally interface{}
	Rescue  bool
}

type Tasks []Task

package stack

import (
	"container/list"
)

var (
	Stacks map[string]*ExecStack = map[string]*ExecStack{}
)

type ExecStack struct {
	Stack *list.List
}

func New(id string) *ExecStack {
	s := &ExecStack{
		Stack: list.New(),
	}
	Stacks[id] = s
	return s
}

/*if you would like to use the object to be pushed and
assign value to it, you must push a pointer
*/
func (s *ExecStack) Push(v interface{}) {
	s.Stack.PushFront(v)
}

func (s *ExecStack) Pop() interface{} {
	top := s.Stack.Front()
	s.Stack.Remove(top)
	return top.Value
}

func (s *ExecStack) GetTop() interface{} {
	if s.GetLen() > 0 {
		top := s.Stack.Front()
		return top.Value
	} else {
		return nil
	}
}

func (s *ExecStack) GetLen() int {
	return s.Stack.Len()
}

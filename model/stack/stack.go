// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package stack

import (
	"container/list"
)

var (
	ExecStack *TaskExecStack
)

type TaskExecStack struct {
	Stack *list.List
}

func init() {
	ExecStack = New()
}

func New() *TaskExecStack {
	return &TaskExecStack{
		Stack: list.New(),
	}
}

func (s *TaskExecStack) Push(v interface{}) {
	s.Stack.PushFront(v)
}

func (s *TaskExecStack) Pop() interface{} {
	top := s.Stack.Front()
	s.Stack.Remove(top)
	return top.Value
}

func (s *TaskExecStack) GetTop() interface{} {
	top := s.Stack.Front()
	return top.Value
}

func (s *TaskExecStack) GetLen() int {
	return s.Stack.Len()
}


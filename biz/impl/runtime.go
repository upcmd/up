// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/stephencheng/up/model/core"
	"github.com/stephencheng/up/model/stack"
)

var (
	TaskerStack = stack.New("tasker")
)

const (
	UP_RUNTIME_TASK_LAYER_NUMBER = "up_runtime_task_layer_number"
)

type TaskRuntimeContext struct {
	Taskname        string
	TasknameLayered string
	ExecbaseVars    *core.Cache
	TaskVars        *core.Cache
	ReturnVars      *core.Cache
}

func TaskerRuntime() *TaskerRuntimeContext {
	return TaskerStack.GetTop().(*TaskerRuntimeContext)
}

func TaskRuntime() *TaskRuntimeContext {
	return TaskerRuntime().Tasker.TaskStack.GetTop().(*TaskRuntimeContext)
}

func SetDryrun() {
	TaskerRuntime().Tasker.Dryrun = true
}

type ExecResult struct {
	Code   int
	Output string
	ErrMsg string
}

type StepRuntimeContext struct {
	Stepname string
	Result   *ExecResult
}

func StepRuntime() *StepRuntimeContext {
	return TaskerRuntime().Tasker.StepStack.GetTop().(*StepRuntimeContext)
}

func StepStack() *stack.ExecStack {
	return TaskerRuntime().Tasker.StepStack
}

type BlockRuntimeContext struct {
	BlockBaseVars *core.Cache
}

func BlockStack() *stack.ExecStack {
	return TaskerRuntime().Tasker.BlockStack
}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package core

import (
	"github.com/stephencheng/up/model/stack"
)

var (
	InstanceName string
	Dryrun       bool
	TaskStack    = stack.New("task")
	StepStack    = stack.New("step")
	TaskBreak    bool
)

const (
	UP_RUNTIME_LAYER_NUMBER = "up_runtime_layer_number"
)

func SetInstanceName(id string) {
	if id != "" {
		InstanceName = id
	} else {
		InstanceName = "nonamed"
	}
}

func SetDryrun() {
	Dryrun = true
}

type TaskRuntimeContext struct {
	Taskname        string
	TasknameLayered string
	ExecbaseVars    *Cache
	TaskVars        *Cache
	ReturnVars      *Cache
}

type StepRuntimeContext struct {
	Stepname string
	Result   *ExecResult
}

type ExecResult struct {
	Code   int
	Output string
	ErrMsg string
}

func TaskRuntime() *TaskRuntimeContext {
	return TaskStack.GetTop().(*TaskRuntimeContext)
}


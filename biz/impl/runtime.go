// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/upcmd/up/model/core"
	"github.com/upcmd/up/model/stack"
	"github.com/upcmd/up/utils"
	u "github.com/upcmd/up/utils"
)

var (
	TaskerStack   = stack.New("tasker")
	UpRunTimeVars = core.NewCache()
	BaseDir       string
)

const (
	UP_RUNTIME_TASK_LAYER_NUMBER    = "up_runtime_task_layer_number"
	UP_RUNTIME_TASKER_LAYER_NUMBER  = "up_runtime_tasker_layer_number"
	UP_RUNTIME_TASK_PIPE_IN_CONTENT = "up_runtime_task_pipe_in_content"
)

type TaskRuntimeContext struct {
	Taskname           string
	TasknameLayered    string
	ExecbaseVars       *core.Cache
	TaskVars           *core.Cache
	ReturnVars         *core.Cache
	IsCalledExternally bool
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

func SetBaseDir(dir string) {
	BaseDir = dir
}

func GetBaseModuleName() string {
	return u.MainConfig.ModuleName
}

type StepRuntimeContext struct {
	Stepname             string
	Result               *u.ExecResult
	ContextVars          *core.Cache
	DataSyncInDvarExpand TransientSyncFunc
}

func StepRuntime() *StepRuntimeContext {
	stack := StepStack()
	if stack.GetLen() > 0 {
		return stack.GetTop().(*StepRuntimeContext)
	} else {
		return nil
	}
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

func ConfigRuntime() *utils.UpConfig {
	return TaskerRuntime().Tasker.Config
}

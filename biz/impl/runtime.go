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
	//accessible only during the deferred finally processing
	UP_RUNTIME_SHELL_EXEC_RESULT = "up_runtime_shell_exec_result"
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

func TaskFinallyStack() *stack.ExecStack {
	return TaskerRuntime().Tasker.FinallyStack
}

func TaskRuntime() *TaskRuntimeContext {
	if TaskerRuntime().Tasker.InFinalExec {
		if TaskFinallyStack() != nil {
			top := TaskFinallyStack().GetTop()
			if top != nil {
				return top.(*TaskRuntimeContext)
			}
		}
	} else {
		if taskStack := TaskerRuntime().Tasker.TaskStack; taskStack != nil {
			top := taskStack.GetTop()
			if top != nil {
				return top.(*TaskRuntimeContext)
			}
		}
	}
	return nil
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
	Timeout              int
	Flags                *[]string
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

func GetVault() *core.Cache {
	return TaskerRuntime().Tasker.SecretVars
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

func debugVars() {
	u.PlnBlue("-debug vars-")

	u.Ppmsg("UpRunTimeVars", UpRunTimeVars)
	u.Ppmsg("RuntimeVarsAndDvarsMerged", TaskerRuntime().Tasker.RuntimeVarsAndDvarsMerged)

	if taskRuntime := TaskRuntime(); taskRuntime != nil {
		u.Ppmsg("ExecbaseVars", taskRuntime.ExecbaseVars)
		u.Ppmsg("TaskVars", taskRuntime.TaskVars)
	}

	if stepRuntime := StepRuntime(); stepRuntime != nil {
		u.Ppmsg("ExecContextVars", stepRuntime.ContextVars)
	} else {
		u.PlnInfo("ExecContextVars is nil")
	}
	u.PlnBlue("--")
}

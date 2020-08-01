// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
)

func runTask(f *CallFuncAction, taskname string) {
	u.PpmsgvvvvvhintHigh(u.Spf("caller's vars to task (%s):", taskname), f.Vars)
	ExecTask(taskname, f.Vars)
}

type CallFuncAction struct {
	Do        interface{}
	Vars      *core.Cache
	Tasknames []string
}

//adapt the abstract step.Do to concrete ShellFuncAction Cmds
func (f *CallFuncAction) Adapt() {
	var taskname string
	var tasknames []string

	switch f.Do.(type) {
	case string:
		taskname = f.Do.(string)
		tasknames = append(tasknames, taskname)

	case []interface{}:
		err := ms.Decode(f.Do, &tasknames)
		u.LogErrorAndPanic("call func adapter", err, "please ref to a task name only")

	default:
		u.LogWarn("call func", "Not implemented or void for no action!")
	}
	f.Tasknames = tasknames
}

func (f *CallFuncAction) Exec() {
	for _, tmptaskname := range f.Tasknames {
		taskname := Render(tmptaskname, f.Vars)
		runTask(f, taskname)
	}
}

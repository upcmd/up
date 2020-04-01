// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
)

func runTask(f *TaskRefFuncAction, taskname string) {
	u.PpmsgvvvvvhintHigh(u.Spf("caller's vars to task (%s):", taskname), f.Vars)
	ExecTask(taskname, f.Vars)
}

type TaskRefFuncAction struct {
	Do   interface{}
	Vars *core.Cache
	Refs []string
}

//adapt the abstract step.Do to concrete ShellFuncAction Cmds
func (f *TaskRefFuncAction) Adapt() {
	var taskname string
	var tasknames []string

	switch f.Do.(type) {
	case string:
		taskname = f.Do.(string)
		tasknames = append(tasknames, taskname)

	case []interface{}:
		err := ms.Decode(f.Do, &tasknames)
		u.LogErrorAndExit("call func adapter", err, "please ref to a task name only")

	default:
		u.LogWarn("call func", "Not implemented or void for no action!")
	}
	f.Refs = tasknames
}

func (f *TaskRefFuncAction) Exec() {
	u.P("calling task:")
	for idx, tmptaskname := range f.Refs {
		taskname := core.Render(tmptaskname, f.Vars)
		u.Pfv("    taskname(%2d):\n%+v\n", idx+1, color.HiBlueString("        \\_%s", taskname))
		runTask(f, taskname)
	}
}


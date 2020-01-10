// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	u "github.com/stephencheng/up/utils"
)

func runTask(f *TaskRefFuncAction, taskname string) {
	ExecTask(taskname)
}

type TaskRefFuncAction struct {
	Do   interface{}
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
		u.LogError("taskref adapter", err)

	default:
		u.P("Not implemented!")
	}
	f.Refs = tasknames
}

func (f *TaskRefFuncAction) Exec() {
	u.P("executing linking tasks")
	for idx, taskname := range f.Refs {
		u.Pfv("    taskname(%2d):\n%+v\n", idx+1, color.HiBlueString("%s", taskname))
		runTask(f, taskname)
	}
}


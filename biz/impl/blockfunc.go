// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
)

func runFlow(f *BlockFuncAction, taskname string) {
	u.PpmsgvvvvvhintHigh(u.Spf("caller's vars to task (%s):", taskname), f.Vars)
	ExecTask(taskname, f.Vars)

	//	RefDir string		//load ref task
	//	refdir := u.CoreConfig.RefDir
	//
	//	if task.Ref != "" {
	//		if task.RefDir != "" {
	//			rawdir := task.RefDir
	//			refdir = core.Render(rawdir, core.RuntimeVarsAndDvarsMerged)
	//		}
	//
	//		rawref := task.Ref
	//		ref := core.Render(rawref, core.RuntimeVarsAndDvarsMerged)
	//
	//		yamlflowroot := u.YamlLoader("flow ref", refdir, ref)
	//		flow := loadRefFlow(yamlflowroot)
	//		(*tasks)[idx].Task = flow
	//	}
	//}

}

type BlockFuncAction struct {
	Do        interface{}
	Vars      *core.Cache
	Tasknames []string
	Steps     *Steps
}

//adapt the abstract step.Do to concrete ShellFuncAction Cmds
func (f *BlockFuncAction) Adapt() {
	var flowname string
	var flow Steps

	switch f.Do.(type) {
	case string:
		//a flow name + refdir to load the flow
		raw := f.Do.(string)
		flowname = core.Render(raw, f.Vars)
		u.P(flowname)

	case []interface{}:
		//detailed steps
		err := ms.Decode(f.Do, &flow)
		u.LogErrorAndExit("load steps", err, "steps has configuration problem, please fix it")

	default:
		u.LogWarn("Block func", "Not implemented or void for no action!")
	}

	f.Steps = &flow
}

func (f *BlockFuncAction) Exec() {
	//for _, step := range *f.Steps {
	//	taskname := core.Render(tmptaskname, f.Vars)
	//	runTask(f, taskname)
	//}
	f.Steps.Exec()
}


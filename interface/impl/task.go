// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stephencheng/up/model"
	"github.com/stephencheng/up/model/cache"
	rt "github.com/stephencheng/up/model/runtime"
	"github.com/stephencheng/up/model/stack"
	u "github.com/stephencheng/up/utils"
	"os"
)

var (
	TaskYmlRoot *viper.Viper
	Tasks       *model.Tasks
	Scopes      *cache.Scopes
)

func InitTasks() {
	TaskYmlRoot = u.YamlLoader("Task", u.CoreConfig.TaskDir, u.CoreConfig.TaskFile)
	loadTasks()
	loadScopes()
	loadRuntimeGlobalVars()
	cache.ScopeProfiles.InitContextInstances()
	cache.SetRuntimeVarsMerged(rt.InstanceName)

}

func ListTasks() {

	u.P("-task list")
	for idx, task := range *Tasks {
		u.Pf("  %d %20s: %s \n", idx+1, task.Name, task.Desc)
		u.Ppmsgvvvv(task)
	}
	u.P("-")

}
func ValidateTask(taskname string) {
	rt.SetDryrun()
	ExecTask(taskname, nil)
}

func ExecTask(taskname string, callerVars *cache.Cache) {
	found := false
	for idx, task := range *Tasks {
		if taskname == task.Name {
			u.Pfvvvv("  loacated task-> %d [%s]: %s \n", idx+1, task.Name, task.Desc)
			found = true
			var steps Steps
			err := ms.Decode(task.Task, &steps)
			stack.ExecStack.Push(callerVars)
			u.Pvvvv("Executing task stack layer:", stack.ExecStack.GetLen())
			if stack.ExecStack.GetLen() > 2 {
				u.LogError("Task exec stack layer check", "Too many layers of task executions, please fix your recursive ref-task configurations")
				os.Exit(-1)
			}
			steps.Exec()
			stack.ExecStack.Pop()
			u.LogError("e:", err)
		}
	}

	if !found {
		u.Pferror("Task %s is not defined!", taskname)
		ListTasks()
	}

}

func loadTasks() error {
	tasksData := TaskYmlRoot.Get("tasks")
	var tasks model.Tasks
	err := ms.Decode(tasksData, &tasks)
	Tasks = &tasks
	return err
}

func loadScopes() error {
	scopesData := TaskYmlRoot.Get("scopes")
	var scopes cache.Scopes
	err := ms.Decode(scopesData, &scopes)
	cache.SetScopeProfiles(&scopes)
	return err
}

func loadRuntimeGlobalVars() error {
	varsData := TaskYmlRoot.Get("vars")
	var vars cache.Cache
	err := ms.Decode(varsData, &vars)
	cache.SetRuntimeGlobalVars(&vars)
	return err
}


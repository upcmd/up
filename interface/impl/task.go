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
	"github.com/stephencheng/up/model/stack"
	u "github.com/stephencheng/up/utils"
	"os"
	"strings"
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
	loadRuntimeDvars()
	cache.ScopeProfiles.InitContextInstances()
	cache.SetRuntimeVarsMerged(cache.InstanceName)
	cache.SetRuntimeGlobalMergedWithDvars()

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
	cache.SetDryrun()
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
			u.LogError("e:", err)
			func() {

				rtContext := cache.RuntimeContext{
					Taskname:   taskname,
					CallerVars: callerVars,
				}
				stack.ExecStack.Push(rtContext)
				u.Pvvvv("Executing task stack layer:", stack.ExecStack.GetLen())
				if stack.ExecStack.GetLen() > 2 {
					u.LogError("Task exec stack layer check", "Too many layers of task executions, please fix your recursive ref-task configurations")
					os.Exit(-1)
				}
				steps.Exec()
				stack.ExecStack.Pop()
			}()

		}
	}

	if !found {
		u.Pferror("Task %s is not defined!", taskname)
		ListTasks()
	}

}

///*
//
// */
//func ValidateTasks() {
//	for idx, task := range *Tasks {
//
//	}
//
//}
//
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

func loadRuntimeGlobalVars() {
	varsData := TaskYmlRoot.Get("vars")
	var vars cache.Cache
	err := ms.Decode(varsData, &vars)
	u.LogError("loadRuntimeGlobalVars", err)
	cache.SetRuntimeGlobalVars(&vars)
}

func loadRuntimeDvars() *cache.Dvars {
	dvarsData := TaskYmlRoot.Get("dvars")
	var dvars cache.Dvars
	err := ms.Decode(dvarsData, &dvars)
	//u.Ptmpdebug("check dvars:", dvars)
	u.LogError("loadRuntimeDvars", err)

	var identified bool
	for _, dvar := range dvars {
		if strings.Contains(dvar.Name, "-") {
			identified = true
			u.Pfvvvv("validating dvar name: %s invalid containing '-'", dvar.Name)
		}
	}

	if identified {
		u.LogError("dvar validate", "the dvar name identified above should be fixed before continue")
		os.Exit(-1)
	}

	cache.SetRuntimeGlobalDvars(&dvars)
	return &dvars

}


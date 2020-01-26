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
	u "github.com/stephencheng/up/utils"
	"io/ioutil"
	"os"
	"path"
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
	cache.SetRuntimeVarsMerged(InstanceName)
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
	SetDryrun()
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

				rtContext := TaskRuntimeContext{
					Taskname:   taskname,
					CallerVars: callerVars,
				}

				TaskStack.Push(&rtContext)
				u.Pvvvv("Executing task stack layer:", TaskStack.GetLen())
				if TaskStack.GetLen() > 2 {
					u.LogError("Task exec stack layer check", "Too many layers of task executions, please fix your recursive ref-task configurations")
					os.Exit(-1)
				}
				steps.Exec()
				TaskStack.Pop()
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
	u.Ptmpdebug("check dvars:", dvars)
	u.LogErrorAndExit("loadRuntimeDvars",
		err,
		"You must fix the data type to be string for a dvar value and try again",
	)

	var identified bool
	for idx, dvar := range dvars {
		if strings.Contains(dvar.Name, "-") {
			identified = true
			u.Pfvvvv("validating dvar name: %s invalid containing '-'", dvar.Name)
		}

		if dvar.Ref != "" && dvar.Value != "" {
			u.InvalidAndExit("validating dvar ref and value", "ref and value can not both exist at the same time")
		}

		if dvar.Ref != "" {
			data, err := ioutil.ReadFile(path.Join(u.CoreConfig.TaskDir, dvar.Ref))
			u.LogErrorAndExit("load dvar value from ref file", err, "please fix file loading problem")
			dvars[idx].Value = string(data)
		}
	}

	if identified {
		u.LogError("dvar validate", "the dvar name identified above should be fixed before continue")
		os.Exit(-1)
	}

	u.Ptmpdebug("aaa", dvars)
	cache.SetRuntimeGlobalDvars(&dvars)
	return &dvars

}


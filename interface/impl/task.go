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

	//TODO: refactory of the runtime init after config is loaded to a proper place
	cache.FuncMapInit()
	loadTasks()
	loadScopes()
	cache.ScopeProfiles.InitContextInstances()
	loadRuntimeGlobalVars()
	loadRuntimeDvars()
	cache.SetRuntimeVarsMerged(InstanceName)
	cache.SetRuntimeGlobalMergedWithDvars()
	//t.ListAllFuncs()
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
			u.Pfvvvv("  located task-> %d [%s]: %s \n", idx+1, task.Name, task.Desc)
			found = true
			var steps Steps
			err := ms.Decode(task.Task, &steps)
			u.LogErrorAndExit("decode steps:", err, "please fix data type in yaml config")
			func() {
				//step name validation
				invalidNames := []string{}
				for _, step := range steps {
					if strings.Contains(step.Name, "-") {
						invalidNames = append(invalidNames, step.Name)
					}
				}

				if len(invalidNames) > 0 {
					u.InvalidAndExit(u.Spf("validating step name fails: %s ", invalidNames), "task name can not contain '-', please use '_' instead, failed names:")
				}
			}()

			func() {
				rtContext := TaskRuntimeContext{
					Taskname:   taskname,
					CallerVars: callerVars,
				}

				TaskStack.Push(&rtContext)
				u.Pvvvv("Executing task stack layer:", TaskStack.GetLen())
				if TaskStack.GetLen() > 2 {
					u.LogError("Task exec stack layer check", "Too many layers of task executions, please fix your recursive .nv-task configurations")
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

func loadTasks() error {
	tasksData := TaskYmlRoot.Get("tasks")
	var tasks model.Tasks
	err := ms.Decode(tasksData, &tasks)
	Tasks = &tasks

	func() {
		//validation
		invalidNames := []string{}
		for idx, task := range *Tasks {
			if strings.Contains(task.Name, "-") {
				invalidNames = append(invalidNames, task.Name)
			}

			//u.Ptmpdebug("99", task)
			if task.Task != nil && task.Ref != "" {
				u.InvalidAndExit("validate task node and ref", "task and ref can not coexist")
			}

			//load ref task
			if task.Ref != "" {
				yamlflowroot := u.YamlLoader("flow ref", u.CoreConfig.TaskDir, task.Ref)
				flow := loadRefFlow(yamlflowroot)
				(*Tasks)[idx].Task = flow
			}
		}

		if len(invalidNames) > 0 {
			u.InvalidAndExit(u.Spf("validating task name fails: %s ", invalidNames), "task name can not contain '-', please use '_' instead, failed names:")
		}
	}()

	//u.Ptmpdebug("222", Tasks)

	return err
}

func loadRefFlow(yamlroot *viper.Viper) *Steps {
	flowData := yamlroot.Get("flow")
	var flow Steps
	err := ms.Decode(flowData, &flow)
	u.LogErrorAndExit("load ref flow", err, "flow of the steps has configuration problem, please fix it")
	return &flow
}

func loadScopes() {
	scopesData := TaskYmlRoot.Get("scopes")
	var scopes cache.Scopes
	err := ms.Decode(scopesData, &scopes)
	cache.SetScopeProfiles(&scopes)

	u.LogErrorAndExit("load full scopes", err, "please assess your scope configuration carefully")
	//u.Ptmpdebug("111", scopes)
}

func loadRuntimeGlobalVars() {
	varsData := TaskYmlRoot.Get("vars")
	var vars cache.Cache
	err := ms.Decode(varsData, &vars)
	//u.Ptmpdebug("111", vars)
	u.LogError("loadRuntimeGlobalVars", err)
	cache.SetRuntimeGlobalVars(&vars)
}

func loadRuntimeDvars() *cache.Dvars {
	dvarsData := TaskYmlRoot.Get("dvars")
	var dvars cache.Dvars
	err := ms.Decode(dvarsData, &dvars)
	u.LogErrorAndExit("loadRuntimeDvars",
		err,
		"You must fix the data type to be\n string for a dvar value and try again. possible problems:\nthe name can not be single character 'y' or 'n' ",
	)

	//dvars.ValidateAndLoading()
	cache.SetRuntimeGlobalDvars(&dvars)
	return &dvars

}


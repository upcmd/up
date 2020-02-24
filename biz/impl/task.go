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
	"strconv"
	"strings"
)

var (
	TaskYmlRoot *viper.Viper
	Tasks       *model.Tasks
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
	cache.SetRuntimeVarsMerged(cache.InstanceName)
	cache.SetRuntimeGlobalMergedWithDvars()
	//t.ListAllFuncs()
}

func ListTasks() {

	u.P("-task list")
	maxlen := 0
	for _, task := range *Tasks {
		tasknamelen := len(task.Name)
		if tasknamelen > maxlen {
			maxlen = tasknamelen
		}
	}

	format := "  %4d| %" + u.Spf("%d", maxlen) + "s: %s \n"

	for idx, task := range *Tasks {
		u.Pf(format, idx+1, task.Name, task.Desc)
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
			u.Pfvvvv("  located task-> %d [%s]: \n", idx+1, task.Name)
			u.LogDesc("task", task.Desc)
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
				rtContext := cache.TaskRuntimeContext{
					Taskname:   taskname,
					CallerVars: callerVars,
				}

				cache.TaskStack.Push(&rtContext)
				u.Pvvvv("Executing task stack layer:", cache.TaskStack.GetLen())
				maxLayers, err := strconv.Atoi(u.CoreConfig.MaxCallLayers)
				u.Ptmpdebug("99", maxLayers)
				u.LogErrorAndExit("evaluate max task stack layer", err, "please setup max MaxCallLayers correctly")

				if maxLayers != 0 && cache.TaskStack.GetLen() > maxLayers {
					u.LogError("Task exec stack layer check:", u.Spf("Too many layers of task executions, max allowed(%d), please fix your recursive call", maxLayers))
					os.Exit(-1)
				}

				steps.Exec()
				cache.TaskStack.Pop()
			}()

		}
	}

	if !found {
		u.Pferror("Task %s is not defined!", taskname)
		ListTasks()
	}

}

func validateAndLoadTaskRef(taks *model.Tasks) {
	//validation
	invalidNames := []string{}
	for idx, task := range *taks {
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
			(*taks)[idx].Task = flow
		}
	}

	if len(invalidNames) > 0 {
		u.InvalidAndExit(u.Spf("validating task name fails: %s ", invalidNames), "task name can not contain '-', please use '_' instead, failed names:")
	}
}

func loadRefTasks() {
	tasksRefList := TaskYmlRoot.Get("tasksref")
	if tasksRefList != nil {
		for _, ref := range tasksRefList.([]interface{}) {
			tasksYamlName := ref.(string)
			tasksYmlRoot := u.YamlLoader(tasksYamlName, u.CoreConfig.TaskDir, tasksYamlName)

			var tasks model.Tasks
			tasksData := tasksYmlRoot.Get("tasks")
			err := ms.Decode(tasksData, &tasks)
			u.LogErrorAndExit(u.Spf("decode tasks:%s", tasksYamlName), err, "please fix configuration in tasks yaml file")
			for _, task := range tasks {
				*Tasks = append(*Tasks, task)
			}
		}
	}
}

func loadTasks() error {
	tasksData := TaskYmlRoot.Get("tasks")
	var tasks model.Tasks
	err := ms.Decode(tasksData, &tasks)
	u.LogErrorAndExit("decode tasks:main", err, "please fix configuration in tasks yaml file")
	Tasks = &tasks

	loadRefTasks()
	validateAndLoadTaskRef(Tasks)

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


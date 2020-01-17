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
	u "github.com/stephencheng/up/utils"
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
	cache.ScopeProfiles.InitContextInstances()
	cache.GetRuntimeInstanceVars(rt.InstanceName, cache.Cache{}, cache.Cache{})
}

type Person struct {
	Name string
	Age  int
}

func ListTasks() {

	u.P("-task list")
	for idx, task := range *Tasks {
		u.Pf("  %d %20s: %s \n", idx+1, task.Name, task.Desc)
		u.Ppmsgvvvv(task)
	}
	u.P("-")

}

func ExecTask(taskname string) {
	found := false
	for idx, task := range *Tasks {
		if taskname == task.Name {
			u.Pfvvvv("  loacated task-> %d [%s]: %s \n", idx+1, task.Name, task.Desc)
			found = true
			//spew.Dump(task)
			var steps Steps
			err := ms.Decode(task.Task, &steps)
			steps.Exec()
			//StepsExec(&steps)
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


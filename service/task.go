// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package service

import (
	//"github.com/davecgh/go-spew/spew"
	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stephencheng/up/model"
	u "github.com/stephencheng/up/utils"
)

var (
	TaskYmlRoot *viper.Viper = u.YamlLoader(u.CoreConfig.TaskDir, u.CoreConfig.TaskFile)
	Tasks       model.Tasks
)

func init() {
	LoadTasks()
}

func ListTasks() {
	for idx, task := range Tasks {
		u.Pf("  %d [%s]: %s \n", idx+1, task.Name, task.Desc)
	}

}

func ExecTask(taskname string) {
	found := false
	for idx, task := range Tasks {
		if taskname == task.Name {
			u.Pfvvvv("  loacated task-> %d [%s]: %s \n", idx+1, task.Name, task.Desc)
			found = true
			//spew.Dump(task)
			var taskImpl model.TaskImpl
			err := ms.Decode(task.Task, &taskImpl)

			u.LogError("e:", err)
			//spew.Dump(taskImpl)
			for idx, step := range taskImpl {
				u.Pfvvvv("  step(%3d): %s\n", idx+1, u.PP(step))
				//u.Pfvvvv("%+v | length: %d\n", step.Do, len(step.Do.([]interface{})))

				cmdCnt := len(step.Do.([]interface{}))
				//u.P("cmd count:", cmdCnt)
				if cmdCnt > 1 {
					var cmds model.ShellCmds
					err = ms.Decode(step.Do, &cmds)
					u.LogError("e:", err)
					for idx, cmd := range cmds {
						u.Pfv("    cmd(%2d): %+v\n", idx+1, cmd)
						u.P("      exec result:", u.MockRunCmd(cmd))
					}
				} else {
					var cmd string
					err = ms.Decode(step.Do, &cmd)
					u.LogError("err:", err)
					u.P("      exec result:", u.MockRunCmd(cmd))
				}
			}
		}
	}

	if !found {
		u.P("Task is not defined!")
	}

}

func LoadTasks() error {
	tasksData := TaskYmlRoot.Get("tasks")
	err := ms.Decode(tasksData, &Tasks)
	return err
}


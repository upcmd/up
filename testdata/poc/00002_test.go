// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package poc

import (
	"github.com/davecgh/go-spew/spew"
	ms "github.com/mitchellh/mapstructure"
	"github.com/stephencheng/up/interface/impl/funcs"
	"github.com/stephencheng/up/model"
	"github.com/stephencheng/up/service"
	td "github.com/stephencheng/up/testdata"
	tl "github.com/stephencheng/up/tests"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test00002(t *testing.T) {
	assert := assert.New(t)
	root := tl.MockLoadYml(tl.TestDataDir, "00001")
	tasks := root.Get("tasks")

	var data model.Tasks
	err := ms.Decode(tasks, &data)
	t.Log("err:", err)
	spew.Dump(data)

	for idx, task := range data {
		t.Logf("task=> %d -> %+v", idx, task)
		t.Logf("%+v", task.Task)

		if task.Name == "task1" {

			//parse task impl
			assert.Equal(task.Name, "task1", "task1 name should be task1")
			var taskImpl service.Steps
			err = ms.Decode(task.Task, &taskImpl)
			t.Log("err:", err)
			spew.Dump(taskImpl)
			for idx, step := range taskImpl {
				t.Logf("step => %d -> %+v", idx, step)
				t.Logf("%+v | length: %d", step.Do, len(step.Do.([]interface{})))
				cmdCnt := len(step.Do.([]interface{}))
				assert.Equal(cmdCnt, 2, "there should be 2 commands defined")
				if cmdCnt > 1 {
					var cmds funcs.ShellCmds
					err = ms.Decode(step.Do, &cmds)
					t.Log("err:", err)
					for idx, cmd := range cmds {
						t.Logf("cmd => %d -> %+v", idx, cmd)
						result := td.RunCmd(t, cmd)
						if idx == 1 {
							assert.Equal("world", strings.Trim(result, "\n"), "command run should return world")
						}
					}
					spew.Dump(cmds)
				} else {
					var cmd string
					err = ms.Decode(step.Do, &cmd)
					t.Log("err:", err)
					spew.Dump(cmd)
					td.RunCmd(t, cmd)
				}
			}

		}
	}

}


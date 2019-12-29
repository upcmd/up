// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package funcs

import (
	//ms "github.com/mitchellh/mapstructure"
	u "github.com/stephencheng/up/utils"
	"os/exec"
)

func runCmd(cmd string) string {
	cmdExec := exec.Command("/bin/sh", "-c", cmd)
	exec.Command("bash", "-c", cmd)
	cmdOutput, err := cmdExec.Output()
	u.LogError("exec error:", err)
	return string(cmdOutput)
}

type ShellFuncAction struct {
	Do   interface{}
	Cmds []string
}

//adapt the abstract step.Do to concrete ShellFuncAction Cmds
func (f *ShellFuncAction) Adapt() {
	var cmd string
	var cmds []string

	switch f.Do.(type) {
	case string:
		cmd = f.Do.(string)
		cmds = append(cmds, cmd)

	case []string:
		cmds = f.Do.([]string)
	}

	f.Cmds = cmds
}

func (f *ShellFuncAction) Exec() {
	//u.P("shell func execed")

	u.P(f.Cmds)
	for idx, cmd := range f.Cmds {
		u.Pfv("    cmd(%2d): %+v\n", idx+1, cmd)
		u.P("      exec result:", runCmd(cmd))
	}
}


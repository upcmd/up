// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package funcs

import (
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	u "github.com/stephencheng/up/utils"
	"os/exec"
)

func runCmd(f *ShellFuncAction, cmd string) string {
	cmdExec := exec.Command("/bin/sh", "-c", cmd)
	exec.Command("bash", "-c", cmd)

	cmdOutput, err := cmdExec.CombinedOutput()
	var result Result
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.Code = exitError.ExitCode()
			result.ErrMsg = string(cmdOutput)
		}
	} else {
		result.Code = 0
		result.Output = string(cmdOutput)
	}

	f.Result = result
	u.LogError("exec error:", err)
	return string(cmdOutput)
}

type Result struct {
	Code   int
	Output string
	ErrMsg string
}

type ShellFuncAction struct {
	Do     interface{}
	Cmds   []string
	Result Result
}

//adapt the abstract step.Do to concrete ShellFuncAction Cmds
func (f *ShellFuncAction) Adapt() {
	var cmd string
	var cmds []string

	switch f.Do.(type) {
	case string:
		cmd = f.Do.(string)
		cmds = append(cmds, cmd)

	case []interface{}:
		err := ms.Decode(f.Do, &cmds)
		u.LogError("shell adapter", err)

	default:
		u.P("Not implemented!")
	}
	f.Cmds = cmds
}

func (f *ShellFuncAction) Exec() {
	u.P("executing shell commands")
	for idx, cmd := range f.Cmds {
		u.Pfv("    cmd(%2d):\n%+v\n", idx+1, color.HiBlueString("%s", cmd))
		//u.Pf("      exec result:\n%s\n", color.HiGreenString("%s", runCmd(f, cmd)))
		runCmd(f, cmd)
		u.Pfv("%s\n", color.HiGreenString("%s", f.Result.Output))
		if f.Result.Code != 0 {
			u.Pfv("      %s\n", color.RedString("%s", f.Result.ErrMsg))
		}

		u.Pfvvvv("      exec result:\ncode:%s\nerror:%s\n",
			color.YellowString("%d", f.Result.Code),
			color.RedString("%s", f.Result.ErrMsg),
		)
	}
}


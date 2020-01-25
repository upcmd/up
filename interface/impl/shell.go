// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	"github.com/stephencheng/up/model/cache"
	t "github.com/stephencheng/up/model/template"
	u "github.com/stephencheng/up/utils"
	"strings"

	"os/exec"
)

func runCmd(f *ShellFuncAction, cmd string) string {
	cmdExec := exec.Command("/bin/sh", "-c", cmd)
	exec.Command("bash", "-c", cmd)

	if cache.Dryrun {
		u.Pdryrun("in dryrun mode and skipping the actual commands")
		return "dryrun result"
	} else {

		cmdOutput, err := cmdExec.CombinedOutput()
		var result ExecResult
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				result.Code = exitError.ExitCode()
				result.ErrMsg = string(cmdOutput)
			}
		} else {
			result.Code = 0
			result.Output = strings.TrimSpace(string(cmdOutput))
		}

		f.Result = result
		u.LogError("exec error:", err)
		return string(cmdOutput)
	}
}

type ShellFuncAction struct {
	Do     interface{}
	Vars   *cache.Cache
	Cmds   []string
	Result ExecResult
	Step   *Step
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
	for idx, tcmd := range f.Cmds {
		u.Pfv("cmd(%2d):\n%+v\n", idx+1, color.HiBlueString("%s", tcmd))
		//u.Pf("      exec result:\n%s\n", color.HiGreenString("%s", runCmd(f, tcmd)))

		//u.Ptmpdebug("1001", tcmd)
		cmd := t.Render(tcmd, f.Vars)

		runCmd(f, cmd)
		u.Pfv("%s\n", color.HiGreenString("%s", f.Result.Output))
		if f.Result.Code != 0 {
			u.Pfv("      %s\n", color.RedString("%s", f.Result.ErrMsg))
		}

		u.Pfvvvv("exec result: {\n  code:%s\n  error:%s\n}\n\n",
			color.YellowString("%d", f.Result.Code),
			color.RedString("%s", f.Result.ErrMsg),
		)
	}

	//u.Ppmsgvvvv(f.Vars)
	u.Ppmsgvvvv(f.Result)
	f.Step.Result = &f.Result
}


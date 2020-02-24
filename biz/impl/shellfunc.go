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
	u "github.com/stephencheng/up/utils"
	"os"
	"os/exec"
	"strings"
)

func runCmd(f *ShellFuncAction, cmd string) string {
	cmdExec := exec.Command("/bin/sh", "-c", cmd)

	func() {
		//inject the envvars
		cmdExec.Env = os.Environ()
		envvarObjMap := f.Vars.GetPrefixMatched("envvar_")
		for k, v := range *envvarObjMap {
			cmdExec.Env = append(cmdExec.Env, u.Spf("%s=%s", k, v.(string)))
		}
	}()

	if cache.Dryrun {
		u.Pdryrun("in dryrun mode and skipping the actual commands")
		return "dryrun result"
	} else {

		cmdOutput, err := cmdExec.CombinedOutput()
		var result cache.ExecResult
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				result.Code = exitError.ExitCode()
				result.ErrMsg = string(cmdOutput)
			}
		} else {
			result.Code = 0
			result.Output = strings.TrimSpace(string(cmdOutput))
		}

		u.SubStepStatus("..", result.Code)

		f.Result = result
		u.LogError("exec error:", err)
		return string(cmdOutput)
	}
}

type ShellFuncAction struct {
	Do     interface{}
	Vars   *cache.Cache
	Cmds   []string
	Result cache.ExecResult
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
		u.LogWarn("shell", "Not implemented or void for no action!")
	}
	f.Cmds = cmds
}

func (f *ShellFuncAction) Exec() {
	//u.P("executing shell commands")
	for idx, tcmd := range f.Cmds {
		u.Pfv("cmd(%2d):\n", idx+1)
		u.Pvv(tcmd)
		cmd := cache.Render(tcmd, f.Vars)
		u.Pfvvvv(" \\_ %+v\n", color.HiBlueString("%s", cmd))
		runCmd(f, cmd)
		u.Pfv("%s\n", color.HiGreenString("%s", f.Result.Output))
		if f.Result.Code != 0 {
			u.Pfv("      %s\n", color.RedString("%s", f.Result.ErrMsg))
		}

		u.Dvvvvv(f.Result)
	}

	cache.StepStack.GetTop().(*cache.StepRuntimeContext).Result = &f.Result
}


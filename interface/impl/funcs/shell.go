// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package funcs

import (
	u "github.com/stephencheng/up/utils"
	"os/exec"
)

func RunCmd(cmd string) string {
	cmdExec := exec.Command("/bin/sh", "-c", cmd)
	exec.Command("bash", "-c", cmd)
	cmdOutput, err := cmdExec.Output()
	u.LogError("exec error:", err)
	return string(cmdOutput)
}

type ShellFuncAction struct {
	Cmds []string
}

func (f *ShellFuncAction) Exec() {

}

type ShellCmds []string


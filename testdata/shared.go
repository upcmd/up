// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package cases

import (
	"os/exec"
	"testing"
)

func RunCmd(t *testing.T, cmd string) string {
	t.Logf("cmd:%s", cmd)
	cmdExec := exec.Command("/bin/sh", "-c", cmd)
	exec.Command("bash", "-c", cmd)
	cmdOutput, err := cmdExec.Output()
	if err != nil {
		panic(err)
	}
	t.Log("exec result:", string(cmdOutput))
	return string(cmdOutput)
}


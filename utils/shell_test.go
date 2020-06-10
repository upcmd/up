// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import "testing"

func TestRunShellCmd(t *testing.T) {
	RunShellCmd("", "ls -lartG")
	RunShellCmd("", "export CM_TEST_XX=haha && env |grep CM_TEST_")
}

func TestRunShellCmdWithEnvVars(t *testing.T) {

}


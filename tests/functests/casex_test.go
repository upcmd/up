// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package functests

import (
	svc "github.com/stephencheng/up/service"
	"github.com/stephencheng/up/tests"
	"testing"
)

func Test00001(t *testing.T) {

	tests.Setup(t)
	svc.InitTasks()
	svc.ListTasks()
	svc.ExecTask("task1")

}

func Test00002(t *testing.T) {
	tests.Setup(t)
	svc.InitTasks()
	svc.ListTasks()
	svc.ExecTask("task1")
}

func Test00003(t *testing.T) {
	tests.Setup(t)
	svc.InitTasks()
	svc.ListTasks()
	svc.ExecTask("task1")
}


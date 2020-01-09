// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package functests

import (
	"github.com/stephencheng/up/interface/impl"
	"github.com/stephencheng/up/tests"
	u "github.com/stephencheng/up/utils"

	"testing"
)

func TestC(t *testing.T) {
	u.InitConfig()

	files := tests.GetUnitTestCollection()
	for _, x := range files {
		u.P("testing:", x)
		tests.Setupx(x)
		impl.InitTasks()
		impl.ListTasks()
		impl.ExecTask("task")
	}
}


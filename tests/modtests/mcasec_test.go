// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package functests

import (
	"github.com/stephencheng/up/biz/impl"
	"github.com/stephencheng/up/tests"
	u "github.com/stephencheng/up/utils"
	"path"
	"testing"
)

func TestC(t *testing.T) {
	dirs := tests.GetModuleTestCollection()

	for _, x := range dirs {
		u.Pln("==testing:", x, "==")
		cfg := tests.SetupMx(path.Join(x))
		t := impl.NewTasker("dev", cfg)
		t.ExecTask("Main", nil, false)
		t.Unset()
	}
}



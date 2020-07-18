// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package functests

import (
	"github.com/upcmd/up/biz/impl"
	"github.com/upcmd/up/tests"
	u "github.com/upcmd/up/utils"
	"os"
	"testing"
)

func init() {
	os.Chdir("../..")
}

func TestC(t *testing.T) {
	cfg := u.NewUpConfig("", "").InitConfig()
	u.MainConfig = cfg
	files := tests.GetUnitTestCollection()
	impl.FuncMapInit()

	for _, x := range files {
		u.Pln("testing:", x)
		u.Pln("work dir:", cfg.AbsWorkDir)
		tests.Setupx(x, cfg)
		t := impl.NewTasker("dev", "", cfg)
		t.ExecTask("task", nil, false)
		t.Unset()
	}
}

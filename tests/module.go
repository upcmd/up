// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package tests

import (
	"github.com/upcmd/up/biz/impl"
	u "github.com/upcmd/up/utils"
	"os"
)

//mock required settings
func SetupMx(dirpath string) *u.UpConfig {
	cfg := u.NewUpConfig(dirpath, "")
	cfg.Secure = &u.SecureSetting{Type: "default_aes", Key: "enc_key"}
	cfg.RefDir = dirpath
	cfg.WorkDir = "refdir"
	cfg.InitConfig()
	u.MainConfig = cfg
	wkdir := cfg.AbsWorkDir
	u.Pln("work dir:", wkdir)
	impl.SetBaseDir(wkdir)
	os.Chdir(wkdir)
	cfg.ShowCoreConfig("moduletest")
	u.Ppmsgvvvvhint("core config", cfg)
	u.Pln(" :test task file:", cfg.TaskFile)
	u.Pln(" :release version:", cfg.Version)
	u.Pln(" :verbose level:", cfg.Verbose)
	return cfg
}

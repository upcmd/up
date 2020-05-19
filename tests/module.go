// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package tests

import (
	u "github.com/stephencheng/up/utils"
	"path"
	"path/filepath"
	"strings"
)

//mock required settings
func SetupMx(filename string, cfg *u.UpConfig) {

	filenameonly := path.Base(filename)

	filenoext := strings.TrimSuffix(filenameonly, filepath.Ext(filenameonly))
	cfg.SetTaskfile(GetTestName(filenoext))
	cfg.SetRefdir("./tests/functests")
	cfg.Secure = &u.SecureSetting{Type: "default_aes", Key: "enc_key"}
	cfg.ShowCoreConfig("mocktest")
	u.Ppmsgvvvvhint("core config", cfg)
	u.Pln(" :test task file:", cfg.TaskFile)
	u.Pln(" :release version:", cfg.Version)
	u.Pln(" :verbose level:", cfg.Verbose)

}


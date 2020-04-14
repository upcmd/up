// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package tests

import (
	"github.com/stephencheng/up/biz/impl"
	m "github.com/stephencheng/up/model"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

//TestHello -> Hello
func GetTestName(testFullName string) string {
	return strings.Replace(testFullName, "Test", "", 1)
}

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

func Setup(prefix string, t *testing.T) {
	u.InitConfig()
	u.CoreConfig.TaskFile = GetTestName(u.Spfv("%s%s", prefix, t.Name()))
	u.ShowCoreConfig()

	u.Pln(" :test task file:", u.CoreConfig.TaskFile)
	u.Pln(" :release version:", u.CoreConfig.Version)
	u.Pln(" :verbose level:", u.CoreConfig.Verbose)
}

func TestT(prefix string, t *testing.T) {
	core.SetInstanceName("dev")
	Setup(prefix, t)
	impl.InitTasks()
	impl.ListTasks()
	impl.ExecTask("task", nil)
}

//mock required settings
func Setupx(filename string) {

	filenameonly := path.Base(filename)

	filenoext := strings.TrimSuffix(filenameonly, filepath.Ext(filenameonly))
	u.CoreConfig.TaskFile = GetTestName(filenoext)
	u.CoreConfig.RefDir = "./tests/functests"
	u.CoreConfig.Secure = &m.SecureSetting{Type: "default_aes", Key: "enc_key"}
	u.ShowCoreConfig()
	u.ShowCoreConfigMsg()
	u.Pln(" :test task file:", u.CoreConfig.TaskFile)
	u.Pln(" :release version:", u.CoreConfig.Version)
	u.Pln(" :verbose level:", u.CoreConfig.Verbose)

}

func GetUnitTestCollection() []string {
	_, filename, _, _ := runtime.Caller(1)
	dir := path.Dir(filename)

	files, err := filepath.Glob(u.Spfv("%s/%s", dir, "c0*.yml"))
	u.LogError("list func test cases", err)

	for _, f := range files {
		u.Pln(f)
	}

	return files
}


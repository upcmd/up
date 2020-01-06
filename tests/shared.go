// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package tests

import (
	svc "github.com/stephencheng/up/service"
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

func Setup(t *testing.T) {
	u.InitConfig()
	u.CoreConfig.TaskFile = GetTestName(u.Spfv("%s%s", "x", t.Name()))
	u.ShowCoreConfig()
	u.P(" :release version:", u.CoreConfig.Version)
	u.P(" :verbose level:", u.CoreConfig.Verbose)
}

func TestT(t *testing.T) {
	Setup(t)
	svc.InitTasks()
	svc.ListTasks()
	svc.ExecTask("task")
}

func Setupx(filename string) {

	filenameonly := path.Base(filename)

	filenoext := strings.TrimSuffix(filenameonly, filepath.Ext(filenameonly))
	u.CoreConfig.TaskFile = GetTestName(filenoext)
	u.ShowCoreConfig()
	u.P(" :release version:", u.CoreConfig.Version)
	u.P(" :verbose level:", u.CoreConfig.Verbose)

}

func GetUnitTestCollection() []string {
	_, filename, _, _ := runtime.Caller(1)
	dir := path.Dir(filename)

	files, err := filepath.Glob(u.Spfv("%s/%s", dir, "c0*.yml"))
	u.LogError("list func test cases", err)

	for _, f := range files {
		u.P(f)
	}

	return files
}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package tests

import (
	"github.com/upcmd/up/biz/impl"
	u "github.com/upcmd/up/utils"
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

func Setup(prefix string, t *testing.T) *u.UpConfig {
	cfg := u.NewUpConfig("", "").InitConfig()
	cfg.SetTaskfile(GetTestName(u.Spfv("%s%s", prefix, t.Name())))
	cfg.ShowCoreConfig("mocktest")

	u.Pln(" :test task file:", impl.ConfigRuntime().TaskFile)
	u.Pln(" :release version:", impl.ConfigRuntime().Version)
	u.Pln(" :verbose level:", impl.ConfigRuntime().Verbose)
	return cfg
}

func TestT(prefix string, t *testing.T) {
	cfg := Setup(prefix, t)
	tasker := impl.NewTasker("dev", cfg)
	tasker.ListTasks()
	tasker.ExecTask("task", nil, false)
}

//mock required settings
func Setupx(filename string, cfg *u.UpConfig) {
	filenameonly := path.Base(filename)
	//filenoext := strings.TrimSuffix(filenameonly, filepath.Ext(filenameonly))
	//cfg.SetTaskfile(GetTestName(filenoext))
	cfg.SetTaskfile(filenameonly)
	cfg.SetRefdir("./tests/functests")
	cfg.Secure = &u.SecureSetting{Type: "default_aes", Key: "enc_key"}
	cfg.ShowCoreConfig("mocktest")
	u.Ppmsgvvvvhint("core config", cfg)
	u.Pln(" :test task file:", cfg.TaskFile)
	u.Pln(" :release version:", cfg.Version)
	u.Pln(" :verbose level:", cfg.Verbose)

}

func GetUnitTestCollection() []string {
	_, filename, _, _ := runtime.Caller(1)
	dir := path.Dir(filename)

	files, err := filepath.Glob(u.Spfv("%s/%s", dir, "c????.yml"))
	u.LogError("list func test cases", err)

	for _, f := range files {
		u.Pln(f)
	}

	return files
}

func GetModuleTestCollection() []string {
	//TODO: to implemene this, delete all below
	_, filename, _, _ := runtime.Caller(1)
	dir := path.Dir(filename)

	files, err := filepath.Glob(u.Spfv("%s/????", dir))
	u.LogError("list func test cases", err)

	for _, f := range files {
		u.Pln(f)
	}

	return files
}

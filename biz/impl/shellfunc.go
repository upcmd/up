// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bytes"
	"context"
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func runCmd(f *ShellFuncAction, cmd string, idx1 int) {
	var result u.ExecResult
	result.Cmd = cmd
	if TaskerRuntime().Tasker.Dryrun {
		u.Pdryrun("in dryrun mode and skipping the actual commands")
		result.Code = 0
		result.Output = strings.TrimSpace("dryrun result")
	} else {
		switch u.MainConfig.ShellType {
		case "GOSH":
			envvarObjMap := f.Vars.GetPrefixMatched("envVar_")
			envVars := map[string]string{}
			for k, v := range *envvarObjMap {
				envVars[k] = v.(string)
			}

			result = u.RunCmd(cmd,
				"",
				&envVars,
			)

			u.Pfv("%s\n", color.HiGreenString("%s", result.Output))

		default:
			defaultTimeout, _ := strconv.Atoi(ConfigRuntime().Timeout)
			timeout := func() time.Duration {
				var t int
				if f.Timeout == 0 {
					t = defaultTimeout
				} else {
					t = f.Timeout
					u.LogWarn("explicit timeout:", u.Spf("%d milli seconds", t))
				}
				return time.Duration(t)
			}()

			ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
			defer cancel()

			cmdExec := exec.CommandContext(ctx, u.MainConfig.ShellType, "-c", cmd)

			func() {
				//inject the envvars
				cmdExec.Env = os.Environ()
				envvarObjMap := f.Vars.GetPrefixMatched("envVar_")
				for k, v := range *envvarObjMap {
					cmdExec.Env = append(cmdExec.Env, u.Spf("%s=%s", k, v.(string)))
				}
			}()

			stdout, err := cmdExec.StdoutPipe()
			if err != nil {
				u.LogError("stdout pipe", err)
			}

			stderr, stderrErr := cmdExec.StderrPipe()
			if err != nil {
				u.LogError("stderr pipe", err)
			}

			if err = cmdExec.Start(); err != nil {
				u.LogError("exec started", err)
			}

			u.PlnInfo("-")
			outputResult := asyncStdReader("stdout", stdout, err, color.HiGreenString, idx1)
			stdErrorResult := asyncStdReader("stderr", stderr, stderrErr, color.HiRedString, idx1)
			u.PlnInfo("\n-")
			err = cmdExec.Wait()

			if ctx.Err() == context.DeadlineExceeded {
				u.LogWarn("timeout", "shell execution timed out")
			}

			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					result.Code = exitError.ExitCode()
					if len(stdErrorResult) > 0 {
						result.ErrMsg = stdErrorResult
					} else {
						result.ErrMsg = err.Error()
					}
				}
			} else {
				result.Code = 0
			}

			result.Output = strings.TrimSpace(outputResult)
			f.Result = result
		}
	}
}

func asyncStdReader(readtype string, ch io.ReadCloser, err error, colorfunc func(format string, a ...interface{}) string, idx1 int) string {
	var result *bytes.Buffer = bytes.NewBufferString("")
	buff := make([]byte, 5120)
	var n int
	for err == nil {
		n, err = ch.Read(buff)
		if n > 0 {
			if !(u.Contains(*StepRuntime().Flags, FLAG_SILENT) && readtype == "stdout") {
				if u.Contains(*StepRuntime().Flags, u.Spf("%s-%d", FLAG_SILENT, idx1)) && readtype == "stdout" {
				} else {
					u.Pfv("%s", colorfunc("%s", string(buff[:n])))
				}
			}
			result.Write(buff[:n])
		}
	}
	return result.String()
}

type ShellFuncAction struct {
	Do      interface{}
	Vars    *core.Cache
	Cmds    []string
	Result  u.ExecResult
	Timeout int
}

//adapt the abstract step.Do to concrete ShellFuncAction Cmds
func (f *ShellFuncAction) Adapt() {
	var cmd string
	var cmds []string
	f.Timeout = StepRuntime().Timeout

	switch f.Do.(type) {
	case string:
		cmd = f.Do.(string)
		cmds = append(cmds, cmd)

	case []interface{}:
		err := ms.Decode(f.Do, &cmds)
		u.LogError("shell adapter", err)

	default:
		u.LogWarn("shell", "Not implemented or void for no action!")
	}
	f.Cmds = cmds
}

func (f *ShellFuncAction) Exec() {
	for idx, tcmd := range f.Cmds {
		u.Pfv("cmd(%2d):\n", idx+1)
		u.Pvv(tcmd)
		cleansed := func() string {
			re := regexp.MustCompile(`{{.*\.secure_.*}}`)
			return re.ReplaceAllString(tcmd, `SECURE_SENSITIVE_INFO_MASKED`)
		}()
		cmd := Render(tcmd, f.Vars)
		cleansedCmd := Render(cleansed, f.Vars)
		u.Pfvvvv("cmd=>:\n%s\n", color.HiBlueString("%s", cleansedCmd))
		runCmd(f, cmd, idx+1)
		u.SubStepStatus("..", f.Result.Code)
		u.Dvvvvv(f.Result)
	}

	StepRuntime().Result = &f.Result
}

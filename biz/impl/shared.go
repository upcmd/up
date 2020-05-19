// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bufio"
	"github.com/fatih/color"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
	ee "github.com/stephencheng/up/utils/error"
	"os"
)

const (
	FUNC_SHELL = "shell"
	FUNC_CALL  = "call"
	FUNC_BLOCK = "block"
	FUNC_CMD   = "cmd"
)

type MustConditionToContinueFunc func() bool

func DryRunOrExit(mark string, mustCondition MustConditionToContinueFunc, conditionDesc string) {

	ok := mustCondition()

	if TaskerRuntime().Tasker.Dryrun {
		color.Green("      %s -> %s", mark, "in dryrun, try to ignore")
		if !ok {
			color.Red("      %s -> %s", mark, "can not continue further due to critical condition not satisfied")
			color.Red("      %s -> %s", mark, conditionDesc)
			os.Exit(-1)
		}
	} else {
		os.Exit(-1)
	}

}

type ContinueFunc func()

//if there is NoFault, then continue
//or if there is a fault in the allowed list, then skip rest, do not run continueFunc
//else the fault is not ignorable, then if use DryRunOrExitFunc
func DryRunAndSkip(mark string, allowedErrors []string, continueFunc ContinueFunc, mustCondition MustConditionToContinueFunc) {
	if mark == ee.NOFAULT {
		continueFunc()
	} else if u.Contains(allowedErrors, mark) {
		//do nothing
		if TaskerRuntime().Tasker.Dryrun {
			u.Pdryrun("in dry run and skip further")
		}
	} else {
		if mustCondition != nil {
			DryRunOrExit("mark", mustCondition, "trying to continue")
		}
	}
}

func pause(execvars *core.Cache) {
	hint := `
enter: continue 
    q: quit
    i: inspect
`
	u.Ppromptvvvvv("pause action to continue", hint)
	reader := bufio.NewReader(os.Stdin)
	keyinput, _ := reader.ReadString('\n')

	switch keyinput {
	case "q\n":
		u.GraceExit("puase action", "client choose to stop continuing the execution")
	case "i\n":
		u.Ppfmsg("runtime exec vars:", *execvars)
		pause(execvars)
	default:
		//continue
	}
}

func IsCalledTask() (called bool) {
	if TaskerStack.GetLen() > 1 {
		called = true
	} else {
		if TaskerRuntime().Tasker.TaskStack.GetLen() > 1 {
			called = true
		} else {
			called = false
		}
	}
	return
}

func IsCalledExternally() (called bool) {
	if TaskerStack.GetLen() > 1 {
		called = true
	} else {
		called = false

	}
	return
}

func IsAtRootTaskLevel() (called bool) {
	if TaskerRuntime().Tasker.TaskStack.GetLen() == 0 {
		called = true
	} else {
		called = false
	}
	return
}


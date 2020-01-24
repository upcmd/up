// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/stephencheng/go-spew/spew"
	rt "github.com/stephencheng/up/model/runtime"
	ee "github.com/stephencheng/up/utils/error"
	"os"

	"runtime"
)

var (
	spewMsgState spew.ConfigState = spew.ConfigState{
		DisableTypes:            true,
		DisableLengths:          true,
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		DisableMethods:          true,
		DisablePointerMethods:   true,
		Indent:                  "  ",
	}
)

func permitted(v string) bool {
	vconfigured := len(CoreConfig.Verbose)
	vallowed := len(v)
	if vconfigured >= vallowed {
		return true
	} else {
		return false
	}
}

func Pvvvv(a ...interface{}) {
	if permitted("vvvv") {
		vvvvv_color_printf("%s\n", fmt.Sprintln(a...))
	}
}

func Pvvvvv(a ...interface{}) {
	if permitted("vvvvv") {
		vvvvv_color_printf("%s\n", fmt.Sprintln(a...))
	}
}

func Dvvvvv(a ...interface{}) {
	if permitted("vvvvv") {
		vvvvv_color_printf("%s\n", spew.Sdump(a...))
	}
}

func Pfdryrun(format string, a ...interface{}) {
	dryrun_color_print(format, a...)
}

func Pdryrun(a ...interface{}) {
	dryrun_color_print("%s\n", a...)
}

func Ppmsgvvvv(a ...interface{}) {
	if permitted("vvvv") {
		msg_color_printf("%s\n", spewMsgState.Sdump(a...))
	}
}

func Sppmsg(a ...interface{}) string {
	return msg_color_sprintf("%s\n", spewMsgState.Sdump(a...))
}

func Ppmsgvvvvhint(hint string, a ...interface{}) {
	Pvvvv(hint)
	Ppmsgvvvv(a...)
}

func Ptmpdebug(mark string, a ...interface{}) {
	if permitted("vvvv") {
		hiColor := color.New(color.FgHiWhite, color.BgRed)
		hiColor.Printf("------%s start-----\n%s\n------%s end-----\n\n", mark, spewMsgState.Sdump(a...), mark)
	}
}

func Pfvvvv(format string, a ...interface{}) {
	if permitted("vvvv") {
		vvvvv_color_printf(format, a...)
	}
}

func Trace() {
	if permitted("vvvvv") {
		pc := make([]uintptr, 15)
		n := runtime.Callers(2, pc)
		frames := runtime.CallersFrames(pc[:n])
		frame, _ := frames.Next()
		fmt.Printf("  \\_%s:%d %s\n", frame.File, frame.Line, frame.Function)
	}
}

func Pfv(format string, a ...interface{}) {
	if permitted("v") {
		fmt.Printf(format, a...)
	}
}

func Pferror(format string, a ...interface{}) {
	verror_color_printf(format, a...)
}

func Spfv(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func LogError(mark string, err interface{}) {
	if err != nil {
		color.Red("      %s -> %s", mark, err)
	}
}

type MustConditionToContinueFunc func() bool

func DryRunOrExit(mark string, mustCondition MustConditionToContinueFunc, conditionDesc string) {

	ok := mustCondition()

	if rt.Dryrun {
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
	} else if Contains(allowedErrors, mark) {
		//do nothing
		if rt.Dryrun {
			dryrun_color_print("in dry run and skip further")
		}
	} else {
		if mustCondition != nil {
			DryRunOrExit("mark", mustCondition, "trying to continue")
		}
	}
}


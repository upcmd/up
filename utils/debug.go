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

func Ppfmsgvvvv(format string, a ...interface{}) {
	if permitted("vvvv") {
		msg_color_printf(format, Spp(a...))
	}
}

func Ppmsgvvvv(a ...interface{}) {
	if permitted("vvvv") {
		msg_color_printf("%s\n", spewMsgState.Sdump(a...))
	}
}

//pretty print
func Spp(a ...interface{}) string {
	str := msg_color_sprintf("%s", spewMsgState.Sdump(a...))

	//fstr1 := strings.Replace(str, ` "`, " ", -1)
	//fstr2 := strings.Replace(fstr1, `\"`, `"`, -1)
	//fstr3 := strings.Replace(fstr2, `",`, ",", -1)
	//fstr4 := strings.Replace(fstr3, `""`, `"`, -1)
	//fstr5 := strings.Replace(fstr4, `"$`, `"`, -1)
	//fstr1 := strings.Replace(str, `\"`, "", -1)
	//fstr2 := strings.Replace(fstr1, `"`, "", -1)
	return str
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


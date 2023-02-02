package utils

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/upcmd/go-spew/spew"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
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
	vconfigured := 3
	if MainConfig != nil {
		vconfigured = len(MainConfig.Verbose)
	}

	vallowed := len(v)
	if vconfigured >= vallowed {
		return true
	} else {
		return false
	}
}

func Pvvvv(a ...interface{}) {
	if permitted("vvvv") {
		vvvv_color_printf("%s\n", fmt.Sprintln(a...))
	}
}

func Pvvv(a ...interface{}) {
	if permitted("vvv") {
		vvvv_color_printf("%s", fmt.Sprintln(a...))
	}
}

func Pvv(a ...interface{}) {
	if permitted("vv") {
		vvvv_color_printf("%s\n", fmt.Sprintln(a...))
	}
}

func Pvvvvv(a ...interface{}) {
	if permitted("vvvvv") {
		vvvv_color_printf("%s", fmt.Sprint(a...))
	}
}

func Dvvvvv(a ...interface{}) {
	if permitted("vvvvv") {
		vvvv_color_printf("%s\n", spew.Sdump(a...))
	}
}

func Dvvvv(a ...interface{}) {
	if permitted("vvvv") {
		vvvv_color_printf("%s\n", spew.Sdump(a...))
	}
}

func PlnInfo(info string) {
	msg_color_printf("%s\n", info)
}

func PlnBlue(info string) {
	blue_color_printf("%s\n", info)
}

func PlnInfoHighlight(info string) {
	hilight_color_printf("%s\n", info)
}

func Pfdryrun(format string, a ...interface{}) {
	dryrun_color_print(format, a...)
}

func Pdryrun(a ...interface{}) {
	dryrun_color_print("%s\n", a...)
}

func Ppmsgvvvv(a ...interface{}) {
	if permitted("vvvv") {
		msg_color_printf("%s", spewMsgState.Sdump(a...))
	}
}

func Ppmsgvvvvv(a ...interface{}) {
	if permitted("vvvvv") {
		msg_color_printf("%s", spewMsgState.Sdump(a...))
	}
}

func Ppmsgvvv(a ...interface{}) {
	if permitted("vvv") {
		msg_color_printf("%s", spewMsgState.Sdump(a...))
	}
}

func Ppmsg(a ...interface{}) {
	msg_color_printf("%s\n", spewMsgState.Sdump(a...))
}

func Ppfmsg(mark string, a ...interface{}) {
	msg_color_printf("%s: %s\n", mark, spewMsgState.Sdump(a...))
}

func PpmsgHintHighPermitted(vlevel string, hint string, a ...interface{}) {
	if permitted(vlevel) {
		vvvv_color_printf("%s:", hint)
		himsg_color_printf("%s\n", spewMsgState.Sdump(a...))
	}
}

func PpmsgvvvvhintHigh(hint string, a ...interface{}) {
	if permitted("vvvv") {
		vvvv_color_printf("%s:", hint)
		himsg_color_printf("%s\n", spewMsgState.Sdump(a...))
	}
}

func PpmsgvvvvvhintHigh(hint string, a ...interface{}) {
	if permitted("vvvvv") {
		vvvv_color_printf("%s:", hint)
		himsg_color_printf("%s\n", spewMsgState.Sdump(a...))
	}
}

func PfHiColor(format string, a ...interface{}) {
	himsg_color_printf(format, a...)
}

func PpmsgvvvhintHigh(hint string, a ...interface{}) {
	if permitted("vvv") {
		vvvv_color_printf("%s:", hint)
		himsg_color_printf("%s\n", spewMsgState.Sdump(a...))
	}
}

func Pprompt(valueName, hint string) {
	hiColor := color.New(color.FgHiWhite, color.BgBlack)
	hiColor.Printf("Enter Value For [%s]: \n%s\n", valueName, hint)

}

func PpmsgvvvvvHigh(hint string, a ...interface{}) {
	if permitted("vvvvv") {
		PpmsgvvvvhintHigh(hint, a...)
	}
}

func PpmsgvvvvHigh(a ...interface{}) {
	if permitted("vvvv") {
		himsg_color_printf("%s\n", spewMsgState.Sdump(a...))
	}
}

func Sppmsg(a ...interface{}) string {
	return msg_color_sprintf("%s\n", spewMsgState.Sdump(a...))
}

func Ppmsgvvvvhint(hint string, a ...interface{}) {
	Pvvvv(hint)
	Ppmsgvvvv(a...)
}

func Ppmsgvvvhint(hint string, a ...interface{}) {
	Pvvv(hint)
	Ppmsgvvv(a...)
}

func Ppmsgvvvvvhint(hint string, a ...interface{}) {
	if permitted("vvvvv") {
		Ppmsgvvvvhint(hint, a...)
	}
}

func Ptmpdebug(mark string, a ...interface{}) {
	hiColor := color.New(color.FgHiWhite, color.BgRed)
	hiColor.Printf("------%s start-----\n%s\n------%s end-----\n\n", mark, spewMsgState.Sdump(a...), mark)
}

func PdebugN(mark int, a ...interface{}) {
	Pdebug(mark, a)
}

func Pdebug(a ...interface{}) {
	hiColor := color.New(FgColorMap[RandomColorName()], BgColorMap[RandomColorName()])
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	loc := fmt.Sprintf("  %s:%d %s", frame.File, frame.Line, frame.Function)
	hiColor.Printf("------start-----\n[%s]\n\n%s\n------end-----\n\n", loc, spewMsgState.Sdump(a...))
}

func PdebugStack(a ...interface{}) {
	Pln("==>")
	Pdebug(a)
	debug.PrintStack()
	Pln("<==")
}

func Pdebugvvvvvvv(a ...interface{}) {
	if permitted("vvvvvv") {
		Pln("==>")
		Pdebug(a)
		Pln("-----trace for reference-----")
		debug.PrintStack()
		Pln("<==")
	}
}

func Ptrace(mark, info string) {
	hiColor := color.New(color.FgHiWhite, color.BgHiBlue)
	hiColor.Printf("%s%s\n", mark, info)
}

func Pfvvvv(format string, a ...interface{}) {
	if permitted("vvvv") {
		vvvv_color_printf(format, a...)
	}
}

func Pfvvvvv(format string, a ...interface{}) {
	if permitted("vvvvv") {
		vvvv_color_printf(format, a...)
	}
}

func PStackTrace() {
	if permitted("vvvvv") {
		Pln("-----trace for reference-----")
		debug.PrintStack()
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

func Pfvvv(format string, a ...interface{}) {
	if permitted("vvv") {
		fmt.Printf(format, a...)
	}
}

func Pfvv(format string, a ...interface{}) {
	if permitted("vv") {
		fmt.Printf(format, a...)
	}
}

func Pferror(format string, a ...interface{}) {
	verror_color_printf(format, a...)
}

func Spfv(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func LogOk(mark string) {
	color.HiGreen("%s ok", mark)
}

func LogDesc(descType string, contextIdx1 int, taskLayerCnt int, name string, desc string) {
	if desc == "" {
		desc = ""
	}

	switch descType {
	case "task":
		color.HiBlue("%sTask%d: [%s: %s ]", strings.Repeat("=", taskLayerCnt), contextIdx1, name, desc)
	case "step":
		if name == "" && desc == "" {
			color.HiBlue("%sStep%d:", strings.Repeat("-", taskLayerCnt), contextIdx1)
		} else {
			if LineCount(desc) > 1 {
				color.HiBlue("%sStep%d: [\n%s%s]", strings.Repeat("-", taskLayerCnt), contextIdx1, name, desc)
			} else {
				color.HiBlue("%sStep%d: [%s: %s ]", strings.Repeat("-", taskLayerCnt), contextIdx1, name, desc)
			}
		}
	case "substep":
		color.HiBlue("%sSubStep%d: [%s: %s ]", strings.Repeat("~", taskLayerCnt), contextIdx1, name, desc)
	}
}

func SubStepStatus(mark string, statusCode int) {
	if statusCode == 0 {
		color.Green(" %s ok", mark)
	} else {
		color.Red(" %s failed(suppressed if it is not the last step)", mark)
	}
}

func LogWarn(mark string, reason string) {
	color.HiYellow(" WARN: [%s] - [%s]", mark, reason)
}

func LogErrorMsg(mark string, reason string) {
	color.Red(" Error must fix: [%s] - [%s]", mark, reason)
}

func LogErrorAndPanic(mark string, err interface{}, hint string) {
	if err != nil {
		color.Red("      %s -> %s", mark, err)
		hiColor := color.New(color.FgHiCyan, color.BgRed)
		hiColor.Printf("ERROR: \n%s\n", hint)
		PStackTrace()
		panic(err.(error).Error())
	}
}

func LogErrorAndExit(mark string, err interface{}, hint string) {
	if err != nil {
		color.Red("      %s -> %s", mark, err)
		hiColor := color.New(color.FgHiCyan, color.BgRed)
		hiColor.Printf("ERROR: \n%s\n", hint)
		os.Exit(-1)
	}
}

func PanicExit(mark string, err interface{}) {
	color.Red("%s -> %s", mark, err)
	PStackTrace()
	os.Exit(-1)
}

func LogError(mark string, err interface{}) {
	if err != nil {
		color.Red("      %s -> %s", mark, err)
		PStackTrace()
	}
}

func LogErrorAndContinue(mark string, err interface{}, hint string) {
	if err != nil {
		color.Red("      %s -> %s", mark, err)
		hiColor := color.New(color.FgHiYellow, color.BgHiMagenta)
		hiColor.Printf("WARN:\n%s\n", hint)

		switch mark {
		case "template rendering":
			PlnInfoHighlight(`trouble shooting tips:
<incompatible types for comparison>: the variable might not be registered, use -v vvv to see the cache, or use inspect cmd to debug
`)
		}

		PStackTrace()
	}
}

func InvalidAndPanic(mark string, hint string) {
	hiColor := color.New(color.FgHiCyan, color.BgRed)
	e := hiColor.Sprintf("  ERROR: %s [%s]\n", mark, hint)
	if TaskPanicCount == 0 && StepPanicCount == 0 {
		Pln(e)
		os.Exit(-255)
	} else {
		panic(e)
	}
}

func GraceExit(mark string, hint string) {
	hiColor := color.New(color.FgHiGreen, color.FgHiWhite)
	hiColor.Printf("  Exit: %s [%s]\n", mark, hint)
	os.Exit(0)
}

func Fail(mark string, hint string) {
	hiColor := color.New(color.BgRed, color.FgHiWhite)
	hiColor.Printf("  Failed: %s [%s]\n", mark, hint)
	os.Exit(-255)
}

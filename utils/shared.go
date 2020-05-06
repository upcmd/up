// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"
)

var (
	P   = fmt.Print
	Pln = fmt.Println
	Pf  = fmt.Printf
	Sp  = fmt.Sprint
	Spf = fmt.Sprintf
)

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func StrIn(s string, aStrList ...string) int {
	for idx, x := range aStrList {
		if x == s {
			return idx
		}
	}
	return -1
}

func CharIsNum(s string) int {
	return StrIn(s, "0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
}

func Sleep(mscnt int) {
	PfHiColor("sleeping %d milli seconds", mscnt)
	total := 0
	for i := 0; i < mscnt; i += 100 {
		Pf("%s", ".")
		total += 100
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(time.Duration(mscnt-total) * time.Millisecond)
	Pln()
}

func PrintContentWithLineNuber(content string) string {
	withLineNuber := ""
	for idx, line := range strings.Split(content, "\n") {
		withLineNuber += fmt.Sprintf("%5d:%s\n", idx+1, line)
	}
	return withLineNuber
}

func DebugYmlContent(dir, filename string) {
	filepath := path.Join(dir, filename)

	content, err := ioutil.ReadFile(filepath)
	LogErrorAndExit(Spf("loading raw content: %s", filepath), err, "please fix file path and name issues")
	LogWarn("Check validity of yml content", PrintContentWithLineNuber(string(content)))
}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"strings"
)

func VVVV(a ...interface{}) {
	if CoreConfig.Verbose == "vvvv" {
		fmt.Println(a...)
	}
}

func Pfvvvv(format string, a ...interface{}) {
	if CoreConfig.Verbose == "vvvv" {
		fmt.Printf(format, a...)
	}
}

func Pfv(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func LogError(mark string, err interface{}) {
	if err != nil {
		color.Red("      %s->%s", mark, err)
	}
}

func PP(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	str := string(s)
	fstr1 := strings.Replace(str, `\"`, "", -1)
	fstr2 := strings.Replace(fstr1, `"`, "", -1)
	return color.YellowString("%s", fstr2)
}

//func PPfvvvv(format string, a interface{}) {
//	if CoreConfig.Verbose == "vvvv" {
//		fmt.Printf(format, PP(a))
//	}
//}
//
//func PPvvvv(a interface{}) {
//	if CoreConfig.Verbose == "vvvv" {
//		fmt.Println(PP(a))
//	}
//}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package template

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	u "github.com/stephencheng/up/utils"
	"path/filepath"
	"runtime"
	"strings"

	//"path/filepath"
	//"runtime"
	//"strings"
	"github.com/leekchan/gtf"
	"text/template"
)

var (
	templateFuncs template.FuncMap
)

func FuncMapInit() {
	taskFuncs := template.FuncMap{
		"OS":   func() string { return runtime.GOOS },
		"ARCH": func() string { return runtime.GOARCH },
		"catLines": func(s string) string {
			s = strings.Replace(s, "\r\n", " ", -1)
			return strings.Replace(s, "\n", " ", -1)
		},
		"splitLines": func(s string) []string {
			s = strings.Replace(s, "\r\n", "\n", -1)
			return strings.Split(s, "\n")
		},
		"fromSlash": func(path string) string {
			return filepath.FromSlash(path)
		},
		"toSlash": func(path string) string {
			return filepath.ToSlash(path)
		},
		"exeExt": func() string {
			if runtime.GOOS == "windows" {
				return ".exe"
			}
			return ""
		},
	}
	//// Deprecated aliases for renamed functions.
	//taskFuncs["FromSlash"] = taskFuncs["fromSlash"]
	//taskFuncs["ToSlash"] = taskFuncs["toSlash"]
	//taskFuncs["ExeExt"] = taskFuncs["exeExt"]

	templateFuncs = sprig.TxtFuncMap()

	for k, v := range gtf.GtfTextFuncMap {
		if _, ok := templateFuncs[k]; !ok {
			//does not exit, add it to funcmap now
			templateFuncs[k] = v
		}
	}

	for k, v := range taskFuncs {
		templateFuncs[k] = v
	}
}

func Render(tstr string, obj interface{}) string {
	tname := "step_item_exec"
	//t, err := template.New(tname).Parse(tstr)
	t, err := template.New(tname).Funcs(templateFuncs).Parse(tstr)
	u.LogErrorAndExit("template rendering", err, "Please fix the template issue and try again")

	var result bytes.Buffer
	//err := t.Execute(&result, obj)
	t.Execute(&result, obj)
	//u.LogError(tname, err)

	return result.String()
}


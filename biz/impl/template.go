// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	u "github.com/stephencheng/up/utils"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
		//--------------------------------------------------------
		"now": func() string {
			t := time.Now()
			return t.Format("2006-01-02T15:04:05+11:00")
		},
		"printobj": func(obj interface{}) string {
			u.Ppmsg(obj)
			return u.Sppmsg(obj)
		},
		//reg do not return any value, so do not expect the dvar value will be something other than empty
		"reg": func(varname string, object interface{}) string {
			TaskRuntime().ExecbaseVars.Put(varname, object)
			return ""
		},
		"dereg": func(varname string) string {
			TaskRuntime().ExecbaseVars.Delete(varname)
			return ""
		},
		"filecontent": func(filepath string) string {
			content, err := ioutil.ReadFile(filepath)
			if err != nil {
				u.LogWarn("filecontent readfile", u.Spf("please fix file path: %s", filepath))
			}
			return string(content)
		},
		"validateMandatoryFailIfNone": func(varname, varvalue string) string {
			if varvalue == "" {
				u.InvalidAndExit("validateMandatoryFailIfNone", u.Spf("Required var:(%s) must not be empty, please fix it", varname))
			}
			return varvalue
		},
	}

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

func ToJson(str string) string {
	return Render("{{toJson .}}", str)
}

func Render(tstr string, obj interface{}) string {
	tname := "."
	t, err := template.New(tname).Funcs(templateFuncs).Parse(tstr)

	u.LogErrorAndExit("template creating", err, "Please fix the template issue and try again")

	var result bytes.Buffer
	err = t.Execute(&result, obj)
	u.LogErrorAndContinue("template rendering problem", err, u.PrintContentWithLineNuber(tstr))

	return result.String()
}

//func init() {
//	FuncMapInit()
//	ListAllFuncs()
//	os.Exit(0)
//}

func ListAllFuncs() {

	var builtins = map[string]string{
		"and":      "and",
		"call":     "call",
		"html":     "HTMLEscaper",
		"index":    "index",
		"js":       "JSEscaper",
		"len":      "length",
		"not":      "not",
		"or":       "or",
		"print":    "fmt.Sprint",
		"printf":   "fmt.Sprintf",
		"println":  "fmt.Sprintln",
		"urlquery": "URLQueryEscaper",

		// Comparisons
		"eq": "eq", // ==
		"ge": "ge", // >=
		"gt": "gt", // >
		"le": "le", // <=
		"lt": "lt", // <
		"ne": "ne", // !=
	}

	for k, v := range builtins {
		u.Pf("%30s : %#v\n", k, v)
	}

	for k, v := range templateFuncs {
		u.Pf("%30s : %#v\n", k, v)
	}

}


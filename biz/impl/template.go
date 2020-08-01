// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/leekchan/gtf"
	"text/template"
)

var (
	templateFuncs template.FuncMap
	taskFuncs     template.FuncMap
)

func safeReg(varname string, object interface{}) string {
	//if this is in dvar processing:
	//need to one way sync the var to the returning var
	TaskRuntime().ExecbaseVars.Put(varname, object)

	//remove this as it will cause dirty data due to dvar processing
	//StepRuntime().ContextVars.Put(varname, object)
	//instead we do a callback to save it to dvar processing scope
	if !StepRuntime().DataSyncInDvarExpand(varname, object) {
		StepRuntime().ContextVars.Put(varname, object)
	}

	return core.ObjToYaml(object)
}

func FuncMapInit() {
	taskFuncs = template.FuncMap{
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
		//--------------------------------------------------------
		"now": func() string {
			t := time.Now()
			return t.Format("2006-01-02T15:04:05+11:00")
		},
		"printObj": func(obj interface{}) string {
			u.Ppmsg(obj)
			return u.Sppmsg(obj)
		},
		"objToYml": func(obj interface{}) string {
			yml := core.ObjToYaml(obj)
			u.PpmsgvvvvvHigh("objToYml", yml)
			return yml
		},
		"ymlToObj": func(yml string) interface{} {
			obj := core.YamlToObj(yml)
			u.PpmsgvvvvvHigh("ymlToObj", obj)
			return obj
		},
		"loopRange": func(start, end int64, regname string) string {
			var looplist []int64 = []int64{}
			for i := start; i <= end; i++ {
				looplist = append(looplist, i)
			}
			safeReg(regname, looplist)
			return regname
		},
		//reg do not return any value, so do not expect the dvar value will be something other than empty
		"reg": safeReg,
		"deReg": func(varname string) string {
			TaskRuntime().ExecbaseVars.Delete(varname)
			return ""
		},
		//keep envExport only in template func as to be carried to any type of func implementation
		"envExport": func(expType, fileToSave string) string {
			var expStr string
			switch expType {
			case "exec_base_env_vars_configured":
				expStr = TaskerRuntime().Tasker.reportContextualEnvVars(TaskRuntime().ExecbaseVars)
			case "exec_env_vars_configured":
				expStr = TaskerRuntime().Tasker.reportContextualEnvVars(StepRuntime().ContextVars)
			}

			if fileToSave != "" {
				ioutil.WriteFile(fileToSave, []byte(expStr), 0644)
			}

			return expStr
		},
		"pathExisted": func(path string) bool {
			pathtstr := u.Spf("{{.%s}}", path)
			return ElementValid(pathtstr, StepRuntime().ContextVars)
		},
		"fileContent": func(filepath string) string {
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				u.LogWarn("fileContent readFile", u.Spf("please fix file path: %s", filepath))
				return ""
			} else {
				content, err := ioutil.ReadFile(filepath)
				if err != nil {
					u.LogWarn("fileContent readFile", u.Spf("please fix file read error, path: %s", filepath))
				}
				return string(content)
			}
		},
		"validateMandatoryFailIfNone": func(varname, varvalue string) string {
			if varvalue == "" {
				u.InvalidAndPanic("validateMandatoryFailIfNone", u.Spf("Required var:(%s) must not be empty, please fix it", varname))
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

	u.LogErrorAndPanic("template creating", err, u.ContentWithLineNumber(tstr))

	var result bytes.Buffer
	err = t.Execute(&result, obj)
	u.LogErrorAndContinue("template rendering", err, u.ContentWithLineNumber(tstr))

	val := result.String()
	if "<no value>" == val {
		val = NONE_VALUE
	}

	return val
}

func ElementValid(path string, obj interface{}) bool {

	t, err := template.New("validator").Funcs(templateFuncs).Parse(path)

	u.LogErrorAndContinue("template element validating", err, "Please fix the template issue and try again")

	var result bytes.Buffer
	err = t.Execute(&result, obj)
	u.LogErrorAndContinue("element validating problem", err, u.ContentWithLineNumber(path))

	if err != nil {
		return false
	} else if result.String() == "<no value>" {
		return false
	} else {
		return true
	}
}

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
func ListUpcmdFuncs() {
	for k, v := range taskFuncs {
		u.Pf("%30s : %#v\n", k, v)
	}

}

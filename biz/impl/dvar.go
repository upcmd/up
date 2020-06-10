// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bufio"
	"github.com/mohae/deepcopy"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Dvars []Dvar

type Dvar struct {
	Name         string
	Value        string
	Desc         string
	Expand       int
	Flags        []string //supported: vvvv, to_object,envvar,
	Rendered     string
	Secure       *u.SecureSetting
	Ref          string
	RefDir       string
	DataKey      string
	DataPath     string
	DataTemplate string
}

func (dvars *Dvars) ValidateAndLoading(contextVars *core.Cache) {
	var identified bool
	for idx, dvar := range *dvars {

		if strings.Contains(dvar.Name, "-") {
			identified = true
			u.InvalidAndExit("validating dvar name", "dvar name can not contain '-', please use '_' instead")
		}
		if u.CharIsNum(dvar.Name[0:1]) != -1 {
			identified = true
			u.InvalidAndExit("validating dvar name", "dvar name can not start with number")
		}

		if dvar.Ref != "" && dvar.Value != "" {
			u.InvalidAndExit("validating dvar ref and value", "ref and value can not both exist at the same time")
		}

		refdir := ConfigRuntime().RefDir
		if dvar.Ref != "" {
			if dvar.RefDir != "" {
				rawdir := dvar.RefDir
				refdir = Render(rawdir, contextVars)
			}

			rawref := dvar.Ref
			ref := Render(rawref, contextVars)

			data, err := ioutil.ReadFile(path.Join(refdir, ref))
			u.LogErrorAndExit("load dvar value from ref file", err, "please fix file loading problem")
			(*dvars)[idx].Value = string(data)
		}
	}

	if identified {
		u.LogError("dvar validate", "the dvar name identified above should be fixed before continue")
		os.Exit(-1)
	}

}

//given a dvars with the vars context, it expands with rendered result
func (dvars *Dvars) Expand(mark string, contextVars *core.Cache) *core.Cache {

	dvars.ValidateAndLoading(contextVars)
	var expandedVars *core.Cache = core.NewCache()

	if *contextVars == nil {
		contextVars = core.NewCache()
	}

	var tmpVars core.Cache = deepcopy.Copy(*contextVars).(core.Cache)
	var tmpDvars Dvars
	tmpDvars = deepcopy.Copy(*dvars).(Dvars)

	var datasource interface{}

	for idx, dvar := range tmpDvars {
		dvarRaw := tmpDvars[idx].Value
		if dvar.Expand == 0 {
			tmpDvars[idx].Expand = 1
		}
		for i := 0; i < tmpDvars[idx].Expand; i++ {
			tval := tmpDvars[idx].Value
			tmpDvars[idx].Value = Render(tval, tmpVars)
		}

		var rval string

		if dvar.DataKey != "" && dvar.DataPath != "" && dvar.DataTemplate != "" {
			u.InvalidAndExit("validating datasource", "datakey, datapath and datatemplate can not coexist at the same time")
		}

		//the rendering using the datakey is the post rendering process
		if dvar.DataKey != "" {
			datakey := Render(dvar.DataKey, tmpVars)
			datasource = tmpVars.Get(datakey)
			rval = Render(dvarRaw, datasource)
		} else {
			rval = tmpDvars[idx].Value
		}

		if dvar.DataPath != "" {
			datapath := Render(dvar.DataPath, tmpVars)
			datasource = core.GetSubObjectFromCache(&tmpVars, datapath, false, ConfigRuntime().Verbose)
			rval = Render(dvarRaw, datasource)
		} else {
			rval = tmpDvars[idx].Value
		}

		if dvar.DataTemplate != "" {
			datatemplate := Render(dvar.DataTemplate, tmpVars)
			datasource = core.YamlToObj(datatemplate)
			rval = Render(dvarRaw, datasource)
		} else {
			rval = tmpDvars[idx].Value
		}

		tmpVars.Put(dvar.Name, rval)
		(*dvars)[idx].Rendered = rval

		if dvar.Name != "void" {
			expandedVars.Put(dvar.Name, rval)
		}

		func() {
			dvar := (*dvars)[idx]
			mergeTarget := &tmpVars
			vlevels := []string{"v", "vv", "vvv", "vvvv", "vvvvv", "vvvvv"}
			if dvar.Flags != nil && len(dvar.Flags) != 0 {

				for _, vlevel := range vlevels {
					if u.Contains(dvar.Flags, vlevel) {
						u.PpmsgHintHighPermitted("v", "dvar> "+dvar.Name, dvar.Rendered)
					}
				}

				if u.Contains(dvar.Flags, "to_object") {
					rawyml := dvar.Rendered

					obj := new(interface{})
					err := yaml.Unmarshal([]byte(rawyml), obj)
					u.LogErrorAndExit("dvar conversion to object:", err, "please validate the ymal content")

					dvarObjName := u.Spf("%s_%s", dvar.Name, "object")
					if dvar.Name != "void" {
						(*mergeTarget).Put(dvarObjName, *obj)
						(*expandedVars).Put(dvarObjName, *obj)
					}

					if TaskerRuntime().Tasker.TaskStack.GetLen() > 0 {
						if u.Contains(dvar.Flags, "reg") {
							if dvar.Name != "void" {
								TaskRuntime().ExecbaseVars.Put(dvarObjName, *obj)
							} else {
								u.LogWarn("?reg a void", "you can't register a object with void name, use a proper name instead or split to multiple steps")
							}
						}
					}

					for _, vlevel := range vlevels {
						if u.Contains(dvar.Flags, vlevel) {
							u.PpmsgHintHighPermitted("v", "dvar> "+dvarObjName, *obj)
						}
					}
				}

				if u.Contains(dvar.Flags, "envvar") {
					envvarName := u.Spf("%s_%s", "envvar", dvar.Name)
					(*mergeTarget).Put(envvarName, dvar.Rendered)
					(*expandedVars).Put(envvarName, dvar.Rendered)
				}

				if TaskerRuntime().Tasker.TaskStack.GetLen() > 0 {
					if u.Contains(dvar.Flags, "reg") {
						if dvar.Name != "void" {
							TaskRuntime().ExecbaseVars.Put(dvar.Name, dvar.Rendered)
						}
					}
				}

				if u.Contains(dvar.Flags, "secure") {
					DecryptAndRegister(ConfigRuntime().Secure, &dvar, mergeTarget, expandedVars)
				}

				if u.Contains(dvar.Flags, "prompt") {
					u.Ppromptvvvvv(dvar.Name, func() string {
						if dvar.Desc != "" {
							return dvar.Desc
						} else {
							return u.Spf("This will be saved as %s's value", dvar.Name)
						}
					}())
					reader := bufio.NewReader(os.Stdin)
					dvarInputValue, _ := reader.ReadString('\n')
					(*mergeTarget).Put(dvar.Name, dvarInputValue)
					(*expandedVars).Put(dvar.Name, dvarInputValue)
				}

				if u.Contains(dvar.Flags, "taskscope") {
					TaskRuntime().TaskVars.Put(dvar.Name, dvar.Rendered)
				}

			}

			if dvar.Secure != nil {
				DecryptAndRegister(ConfigRuntime().Secure, &dvar, mergeTarget, expandedVars)
			}

		}()

	}

	u.Pfvvvvv("[%s] dvar expanded result:\n%s\n", mark, u.Sppmsg(*expandedVars))

	return expandedVars
}



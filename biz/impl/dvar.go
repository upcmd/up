// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bufio"
	"github.com/mohae/deepcopy"
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Dvars []Dvar
type EnvVars []EnvVar

type Dvar struct {
	Name         string
	Value        string
	Desc         string
	Expand       int
	Flags        []string //supported: vvvv, toObj,envVar,
	Rendered     string
	Secure       *u.SecureSetting
	Ref          string
	RefDir       string
	DataKey      string
	DataPath     string
	DataTemplate string
}

type EnvVar struct {
	Name  string
	Value string
}

func (dvars *Dvars) ValidateAndLoading(contextVars *core.Cache) {
	var identified bool
	for idx, dvar := range *dvars {

		if strings.Contains(dvar.Name, "-") {
			identified = true
			u.InvalidAndPanic("validating dvar name", "dvar name can not contain '-', please use '_' instead")
		}
		if u.CharIsNum(dvar.Name[0:1]) != -1 {
			identified = true
			u.InvalidAndPanic("validating dvar name", "dvar name can not start with number")
		}

		if dvar.Ref != "" && dvar.Value != "" {
			u.InvalidAndPanic("validating dvar ref and value", "ref and value can not both exist at the same time")
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
			u.LogErrorAndPanic("load dvar value from ref file", err, "please fix file loading problem")
			(*dvars)[idx].Value = string(data)
		}
	}

	if identified {
		u.InvalidAndPanic("dvar validate", "the dvar name identified above should be fixed before continue")
	}

}

//return false means not in context of dvar scope
type TransientSyncFunc func(key string, val interface{}) bool

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

	//this is to ensure data consistency of the one way return and overriding from dvar expand to step vars(and context vars)
	transientSync := func(key string, val interface{}) bool {
		//ensure all template reg change is carried over
		tmpVars.Put(key, val)
		expandedVars.Put(key, val)
		return true
	}
	transientSyncVoid := func(key string, val interface{}) bool { return false }

	stepRuntime := StepRuntime()
	if stepRuntime != nil {
		stepRuntime.DataSyncInDvarExpand = transientSync
	}

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
			u.InvalidAndPanic("validating datasource", "datakey, datapath and datatemplate can not coexist at the same time")
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

		if rval == "" {
			rval = NONE_VALUE
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
			var dvarObjName string
			var dvarNameKept bool
			var objConverted = new(interface{})
			if dvar.Flags != nil && len(dvar.Flags) != 0 {

				for _, vlevel := range vlevels {
					if u.Contains(dvar.Flags, vlevel) {
						u.PpmsgHintHighPermitted("v", "dvar> "+dvar.Name, dvar.Rendered)
						u.Pln("-")
						u.PlnInfo(dvar.Rendered)
					}
				}

				if u.Contains(dvar.Flags, "toObj") {
					rawyml := dvar.Rendered

					err := yaml.Unmarshal([]byte(rawyml), objConverted)
					u.LogErrorAndPanic("dvar conversion to object:", err, u.ContentWithLineNumber(rawyml))

					dvarObjName = func() (dvarname string) {
						if u.Contains(dvar.Flags, "keepName") {
							dvarname = dvar.Name
							dvarNameKept = true
						} else {
							dvarname = u.Spf("%s_%s", dvar.Name, "object")
						}
						return
					}()

					if dvar.Name != "void" {
						(*mergeTarget).Put(dvarObjName, *objConverted)
						(*expandedVars).Put(dvarObjName, *objConverted)
					}
					if TaskerRuntime().Tasker.TaskStack.GetLen() > 0 {
						if u.Contains(dvar.Flags, "reg") {
							if dvar.Name != "void" {
								TaskRuntime().ExecbaseVars.Put(dvarObjName, *objConverted)
							} else {
								u.LogWarn("?reg a void", "you can't register a object with void name, use a proper name instead or split to multiple steps")
							}
						}
					}

					for _, vlevel := range vlevels {
						if u.Contains(dvar.Flags, vlevel) {
							u.PpmsgHintHighPermitted("v", "dvar[object]> "+dvarObjName, *objConverted)
						}
					}
				}

				if u.Contains(dvar.Flags, "envVar") {
					envvarName := u.Spf("%s_%s", "envVar", dvar.Name)
					(*mergeTarget).Put(envvarName, dvar.Rendered)
					(*expandedVars).Put(envvarName, dvar.Rendered)
					os.Setenv(dvar.Name, dvar.Rendered)
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
					saneValue := u.RemoveCr(dvarInputValue)
					(*mergeTarget).Put(dvar.Name, saneValue)
					(*expandedVars).Put(dvar.Name, saneValue)
					dvar.Rendered = saneValue
				}

				if u.Contains(dvar.Flags, "taskScope") {
					if !dvarNameKept {
						TaskRuntime().TaskVars.Put(dvar.Name, dvar.Rendered)
					}
					if dvarObjName != "" {
						//keepname only applies to toObj case
						if dvarNameKept {
							TaskRuntime().TaskVars.Put(dvar.Name, *objConverted)
						} else {
							TaskRuntime().TaskVars.Put(dvarObjName, *objConverted)
						}
					}
				}
			}

			if dvar.Secure != nil {
				DecryptAndRegister(ConfigRuntime().Secure, &dvar, mergeTarget, expandedVars)
			}

		}()

	}

	u.Pfvvvvv("[%s] dvar expanded result:\n%s\n", mark, u.Sppmsg(*expandedVars))
	if stepRuntime != nil {
		stepRuntime.DataSyncInDvarExpand = transientSyncVoid
	}

	return expandedVars
}

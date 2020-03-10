// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package core

import (
	"github.com/mohae/deepcopy"
	"github.com/stephencheng/up/model"
	u "github.com/stephencheng/up/utils"
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
	Secure       *model.SecureSetting
	Ref          string
	RefDir       string
	DataKey      string
	DataPath     string
	DataTemplate string
}

func (dvars *Dvars) ValidateAndLoading(contextVars *Cache) {
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

		refdir := u.CoreConfig.RefDir
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
func (dvars *Dvars) Expand(mark string, contextVars *Cache) *Cache {

	dvars.ValidateAndLoading(contextVars)
	var expandedVars *Cache = NewCache()

	if *contextVars == nil {
		contextVars = NewCache()
	}

	var tmpVars Cache = deepcopy.Copy(*contextVars).(Cache)
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
			datasource = GetSubObjectFromCache(&tmpVars, datapath, false)
			rval = Render(dvarRaw, datasource)
		} else {
			rval = tmpDvars[idx].Value
		}

		if dvar.DataTemplate != "" {
			datatemplate := Render(dvar.DataTemplate, tmpVars)
			datasource = YamlToObj(datatemplate)
			rval = Render(dvarRaw, datasource)
		} else {
			rval = tmpDvars[idx].Value
		}

		tmpVars.Put(dvar.Name, rval)
		(*dvars)[idx].Rendered = rval

		if dvar.Name != "void" {
			expandedVars.Put(dvar.Name, rval)
		}
	}

	u.Pfvvvvv("[%s] dvar expanded result:\n%s\n", mark, u.Sppmsg(*expandedVars))

	return expandedVars
}


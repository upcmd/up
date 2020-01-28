// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package cache

import (
	"github.com/mohae/deepcopy"
	t "github.com/stephencheng/up/model/template"
	u "github.com/stephencheng/up/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Dvars []Dvar

type Dvar struct {
	Name     string
	Value    string
	Desc     string
	Expand   int
	Flags    []string
	Rendered string
	Ref      string
}

func (dvars *Dvars) ValidateAndLoading() {
	var identified bool
	for idx, dvar := range *dvars {
		if strings.Contains(dvar.Name, "-") {
			identified = true
			u.InvalidAndExit("validating dvar name", "dvar name can not contain '-', please use '_' instead")
		}

		if dvar.Ref != "" && dvar.Value != "" {
			u.InvalidAndExit("validating dvar ref and value", "ref and value can not both exist at the same time")
		}

		if dvar.Ref != "" {
			data, err := ioutil.ReadFile(path.Join(u.CoreConfig.TaskDir, dvar.Ref))
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
func (dvars *Dvars) Expand(mark string, vars *Cache) *Cache {

	dvars.ValidateAndLoading()
	var expandedVars *Cache = New()

	var tmpDvars Dvars
	tmpDvars = deepcopy.Copy(*dvars).(Dvars)
	var tmpVars Cache = deepcopy.Copy(*vars).(Cache)

	for idx, dvar := range tmpDvars {
		if dvar.Expand == 0 {
			tmpDvars[idx].Expand = 1
		}
		for i := 0; i < tmpDvars[idx].Expand; i++ {
			tval := tmpDvars[idx].Value
			tmpDvars[idx].Value = t.Render(tval, tmpVars)
		}

		rval := tmpDvars[idx].Value
		tmpVars.Put(dvar.Name, rval)
		(*dvars)[idx].Rendered = rval
		expandedVars.Put(dvar.Name, rval)
	}

	u.Pfvvvv("[%s] dvar expanded result:\n%s\n", mark, u.Sppmsg(*expandedVars))

	return expandedVars
}


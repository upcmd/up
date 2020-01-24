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
)

type Dvars []Dvar

type Dvar struct {
	Name   string
	Value  string
	Desc   string
	Expand int
}

func (dvars *Dvars) Expand(vars *Cache) *Cache {
	var expandedVars *Cache = New()

	var tmpDvars Dvars
	//u.Ptmpdebug("xxx", dvars)

	tmpDvars = deepcopy.Copy(*dvars).(Dvars)
	var tmpVars Cache = deepcopy.Copy(*vars).(Cache)

	for idx, dvar := range tmpDvars {
		if dvar.Expand == 0 {
			tmpDvars[idx].Expand = 1
		}
		//u.Ptmpdebug("dvar:", idx+1, tmpDvars[idx])

		for i := 0; i < tmpDvars[idx].Expand; i++ {

			tval := tmpDvars[idx].Value
			//u.Ptmpdebug("1111", tval)
			tmpDvars[idx].Value = t.Render(tval, tmpVars)
			//u.Ptmpdebug("2222", tmpDvars[idx].Value)
		}

		rval := tmpDvars[idx].Value
		tmpVars.Put(dvar.Name, rval)
		expandedVars.Put(dvar.Name, rval)
	}

	//u.Ptmpdebug("dvar expanded result:", *expandedVars)
	u.Ppmsgvvvvhint("dvar expanded result", *expandedVars)

	return expandedVars
}

func (dvars *Dvars) MergeTo(vars *Cache) {

}


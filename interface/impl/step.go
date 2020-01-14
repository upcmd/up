// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	ic "github.com/stephencheng/up/interface"
	u "github.com/stephencheng/up/utils"
)

type Step struct {
	Do   interface{} //FuncImpl
	Func string
	Desc string
}

func (step *Step) Exec() {
	var action ic.Do

	switch step.Func {

	case FUNC_SHELL:
		funcAction := ShellFuncAction{
			Do: step.Do,
		}
		action = ic.Do(&funcAction)

	case FUNC_TASK_REF:
		funcAction := TaskRefFuncAction{
			Do: step.Do,
		}
		action = ic.Do(&funcAction)

	}

	action.Adapt()
	action.Exec()

}

type Steps []Step

func (steps *Steps) Exec() {

	for idx, step := range *steps {
		u.Pfvvvv("  step(%3d): %s\n", idx+1, u.Spp(step))
		//u.Pfvvvv("%+v | length: %d\n", step.Do, len(step.Do.([]interface{})))
		step.Exec()
	}

}


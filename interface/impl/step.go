// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/imdario/mergo"
	ic "github.com/stephencheng/up/interface"
	"github.com/stephencheng/up/model/cache"
	ee "github.com/stephencheng/up/utils/error"

	u "github.com/stephencheng/up/utils"
)

type Step struct {
	Name string
	Do   interface{} //FuncImpl
	Func string
	Vars *cache.Cache
	Desc string
	Reg  string
}

//this is final merged exec vars the individual step will use
//this step will merge the vars with the caller's stack vars
func getExecVars(funcname string, stepVars *cache.Cache) *cache.Cache {
	vars := cache.GetRuntimeExecVars(funcname, stepVars)
	callerVars := TaskStack.GetTop().(*TaskRuntimeContext).CallerVars
	//u.Ptmpdebug("callerVars", callerVars)

	if callerVars != nil {
		mergo.Merge(vars, callerVars, mergo.WithOverride)
	}
	//u.Ptmpdebug("exec vars", vars)
	return vars
}

func (step *Step) Exec() {
	var action ic.Do
	//u.Ptmpdebug("step debug", step)

	var bizErr *ee.Error = ee.New()

	switch step.Func {

	case FUNC_SHELL:
		funcAction := ShellFuncAction{
			Do:   step.Do,
			Vars: getExecVars(FUNC_SHELL, step.Vars),
		}
		action = ic.Do(&funcAction)

	case FUNC_TASK_REF:
		funcAction := TaskRefFuncAction{
			Do: step.Do,
			//TODO: see if we should allow recursive call
			//Vars: cache.GetRuntimeExecVars(FUNC_TASK_REF, step.Vars),
			Vars: getExecVars(FUNC_TASK_REF, step.Vars),
		}
		action = ic.Do(&funcAction)

	default:
		u.LogError("Step dispatch", "func name is not recognised and implemented")
		bizErr.Mark = "func name not implemented"
	}

	//example to stop further steps
	//f := u.MustConditionToContinueFunc(func() bool {
	//	return action != nil
	//})
	//
	//u.DryRunOrExit("Step Exec", f, "func name must be valid")

	alloweErrors := []string{
		"func name not implemented",
	}

	DryRunAndSkip(
		bizErr.Mark,
		alloweErrors,
		ContinueFunc(
			func() {
				action.Adapt()
				action.Exec()
			}),
		nil,
	)

}

type Steps []Step

func (steps *Steps) Exec() {

	for idx, step := range *steps {
		u.Pfvvvv("  step(%3d): %s\n", idx+1, u.Sppmsg(step))

		func() {
			rtContext := StepRuntimeContext{
				Stepname: step.Name,
			}
			StepStack.Push(&rtContext)
			step.Exec()

			result := StepStack.GetTop().(*StepRuntimeContext).Result
			taskname := TaskStack.GetTop().(*TaskRuntimeContext).Taskname
			if step.Reg == "auto" {
				cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("register_%s_%s", taskname, step.Name), result.Output)
			} else if step.Reg != "" {
				cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("register_%s_%s", taskname, step.Reg), result.Output)
			}

			StepStack.Pop()
		}()

	}

	u.Ptmpdebug("register test", cache.RuntimeVarsAndDvarsMerged)

}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/imdario/mergo"
	"github.com/mohae/deepcopy"
	ic "github.com/stephencheng/up/interface"
	"github.com/stephencheng/up/model/cache"
	ee "github.com/stephencheng/up/utils/error"

	u "github.com/stephencheng/up/utils"
)

type Step struct {
	Name  string
	Do    interface{} //FuncImpl
	Func  string
	Vars  cache.Cache
	Dvars cache.Dvars
	Desc  string
	Reg   string
}

//this is final merged exec vars the individual step will use
//this step will merge the vars with the caller's stack vars
func (step *Step) GetExecVarsWithRefOverrided(funcname string) *cache.Cache {
	vars := step.getRuntimeExecVars(funcname)
	callerVars := TaskStack.GetTop().(*TaskRuntimeContext).CallerVars
	//u.Ptmpdebug("callerVars", callerVars)

	if callerVars != nil {
		mergo.Merge(vars, callerVars, mergo.WithOverride)
	}
	u.Ppmsgvvvvhint("overall final exec vars:", vars)
	return vars
}

/*
merge localvars to above RuntimeVarsAndDvarsMerged to get final runtime exec vars
the localvars is the vars in the step
*/
func (step *Step) getRuntimeExecVars(mark string) *cache.Cache {
	var execvars cache.Cache

	execvars = deepcopy.Copy(*cache.RuntimeVarsAndDvarsMerged).(cache.Cache)

	if step.Vars != nil {
		mergo.Merge(&execvars, step.Vars, mergo.WithOverride)

		u.Pfvvvv("current exec runtime[%s] vars:", mark)
		u.Ppmsgvvvv(execvars)
		u.Dvvvvv(execvars)
	}

	localVarsMergedWithDvars := cache.VarsMergedWithDvars("local", &step.Vars, &step.Dvars, &execvars)

	if localVarsMergedWithDvars.Len() > 0 {
		mergo.Merge(&execvars, localVarsMergedWithDvars, mergo.WithOverride)
	}

	return &execvars
}

func (step *Step) Exec() {
	var action ic.Do
	//u.Ptmpdebug("step debug", step)

	var bizErr *ee.Error = ee.New()

	switch step.Func {

	case FUNC_SHELL:
		funcAction := ShellFuncAction{
			Do:   step.Do,
			Vars: step.GetExecVarsWithRefOverrided(FUNC_SHELL),
		}
		action = ic.Do(&funcAction)

	case FUNC_TASK_REF:
		funcAction := TaskRefFuncAction{
			Do:   step.Do,
			Vars: step.GetExecVarsWithRefOverrided(FUNC_TASK_REF),
		}
		action = ic.Do(&funcAction)

	case FUNC_NOOP:
		funcAction := NoopFuncAction{
			Do:   step.Do,
			Vars: step.GetExecVarsWithRefOverrided(FUNC_NOOP),
		}
		action = ic.Do(&funcAction)

	default:
		//u.LogError("Step dispatch", "func name is not recognised and implemented")
		u.InvalidAndExit("Step dispatch", "func name is not recognised and implemented")
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
		u.Pf("step(%3d):\n", idx+1)
		//u.Pfvvvv("  step(%3d): %s\n", idx+1, u.Sppmsg(step))
		u.Ppmsgvvvv(step)

		func() {
			rtContext := StepRuntimeContext{
				Stepname: step.Name,
			}
			StepStack.Push(&rtContext)
			step.Exec()

			result := StepStack.GetTop().(*StepRuntimeContext).Result
			taskname := TaskStack.GetTop().(*TaskRuntimeContext).Taskname
			if u.Contains([]string{FUNC_SHELL, FUNC_TASK_REF}, step.Func) {
				if step.Reg == "auto" {
					cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("register_%s_%s", taskname, step.Name), result.Output)
				} else if step.Reg != "" {
					cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("register_%s_%s", taskname, step.Reg), result.Output)
				} else {
					if step.Func == FUNC_SHELL {
						cache.RuntimeVarsAndDvarsMerged.Put("last_task_result", result.Output)
					}
				}
			}
			StepStack.Pop()
		}()

	}

}


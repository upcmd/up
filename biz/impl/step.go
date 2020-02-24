// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bufio"
	"github.com/imdario/mergo"
	"github.com/mohae/deepcopy"
	"github.com/stephencheng/up/biz"
	"github.com/stephencheng/up/model/cache"
	u "github.com/stephencheng/up/utils"
	ee "github.com/stephencheng/up/utils/error"
	"os"

	//"gopkg.in/yaml.v2"
	"reflect"
	"strconv"
)

type Step struct {
	Name  string
	Do    interface{} //FuncImpl
	Func  string
	Vars  cache.Cache
	Dvars cache.Dvars
	Desc  string
	Reg   string
	Flags []string //ignore_error |
	If    string
	//Loop  *[]interface{}
	Loop interface{}
}

type Steps []Step

//this is final merged exec vars the individual step will use
//this step will merge the vars with the caller's stack vars
func (step *Step) GetExecVarsWithRefOverrided(funcname string) *cache.Cache {
	vars := step.getRuntimeExecVars(funcname)
	callerVars := cache.TaskStack.GetTop().(*cache.TaskRuntimeContext).CallerVars
	//u.Ptmpdebug("callerVars", callerVars)

	if callerVars != nil {
		mergo.Merge(vars, callerVars, mergo.WithOverride)
	}

	u.Ppmsgvvvhint("overall final exec vars:", vars)
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

type LoopItem struct {
	Index  int
	Index1 int
	Item   interface{}
}

func chainAction(action *biz.Do) {
	(*action).Adapt()
	(*action).Exec()
}

func (step *Step) Exec() {
	var action biz.Do
	//u.Ptmpdebug("step debug", step)

	var bizErr *ee.Error = ee.New()
	var stepExecVars *cache.Cache
	stepExecVars = step.GetExecVarsWithRefOverrided("get plain exec vars")

	routeFuncType := func(loopItem *LoopItem) {
		if loopItem != nil {
			stepExecVars.Put("loopitem", loopItem.Item)
			stepExecVars.Put("loopindex", loopItem.Index)
			stepExecVars.Put("loopindex1", loopItem.Index1)
		}

		switch step.Func {
		case FUNC_SHELL:
			funcAction := ShellFuncAction{
				Do:   step.Do,
				Vars: stepExecVars,
			}
			action = biz.Do(&funcAction)

		case FUNC_CALL:
			funcAction := TaskRefFuncAction{
				Do:   step.Do,
				Vars: stepExecVars,
			}
			action = biz.Do(&funcAction)

		case FUNC_CMD:
			funcAction := CmdFuncAction{
				Do:   step.Do,
				Vars: stepExecVars,
			}
			action = biz.Do(&funcAction)

		default:
			u.InvalidAndExit("Step dispatch", "func name is not recognised and implemented")
			bizErr.Mark = "func name not implemented"
		}
	}

	dryRunOrContinue := func() {
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
					if step.Loop != nil {
						func() {
							//loop points to a var name which is a slice
							if reflect.TypeOf(step.Loop).Kind() == reflect.String {
								loopVarName := cache.Render(step.Loop.(string), stepExecVars)
								loopObj := stepExecVars.Get(loopVarName)
								if loopObj == nil {
									u.InvalidAndExit("Evaluating loop var and object", u.Spf("Please use a correct varname:(%s) containing a list of values", loopVarName))
								}
								if reflect.TypeOf(loopObj).Kind() == reflect.Slice {

									switch loopObj.(type) {
									case []interface{}:
										for idx, item := range loopObj.([]interface{}) {
											routeFuncType(&LoopItem{idx, idx + 1, item})
											chainAction(&action)
										}

									case []string:
										for idx, item := range loopObj.([]string) {
											routeFuncType(&LoopItem{idx, idx + 1, item})
											chainAction(&action)
										}

									default:
										u.LogWarn("loop item evaluation", "Loop item type is not supported yet!")
									}

								} else {
									u.InvalidAndExit("evaluate loop var", "loop var is not a array/list/slice")
								}
							} else if reflect.TypeOf(step.Loop).Kind() == reflect.Slice {
								//loop itself is a slice
								for idx, item := range step.Loop.([]interface{}) {
									routeFuncType(&LoopItem{idx, idx + 1, item})
									chainAction(&action)
								}

							} else {
								u.InvalidAndExit("evaluate loop items", "please either use a list or a template evaluation which could result in a value of a list")
							}
						}()

					} else {
						routeFuncType(nil)
						chainAction(&action)

					}

				}),
			nil,
		)
	}

	func() {
		if step.If != "" {
			IfEval := cache.Render(step.If, stepExecVars)
			goahead, err := strconv.ParseBool(IfEval)
			u.LogErrorAndExit("evaluate condition", err, u.Spf("please fix if condition evaluation: [%s]", IfEval))
			if goahead {
				dryRunOrContinue()
			} else {
				u.Pvvvv("condition failed, skip executing step", step.Name)
			}
		} else {
			dryRunOrContinue()
		}

	}()

}

func (steps *Steps) Exec() {

	for idx, step := range *steps {
		u.Pf("step(%3d):\n", idx+1)
		//u.Pfvvvv("  step(%3d): %s\n", idx+1, u.Sppmsg(step))
		u.LogDesc("step", step.Desc)
		u.Ppmsgvvvv(step)

		execStep := func() {
			rtContext := cache.StepRuntimeContext{
				Stepname: step.Name,
				//Flags:    &step.Flags,
			}
			cache.StepStack.Push(&rtContext)

			step.Exec()

			result := cache.StepStack.GetTop().(*cache.StepRuntimeContext).Result
			taskname := cache.TaskStack.GetTop().(*cache.TaskRuntimeContext).Taskname

			if u.Contains([]string{FUNC_SHELL, FUNC_CALL}, step.Func) {
				if step.Reg == "auto" {
					cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("register_%s_%s", taskname, step.Name), result.Output)
				} else if step.Reg != "" {
					cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("%s", step.Reg), result.Output)
				} else {
					if step.Func == FUNC_SHELL {
						cache.RuntimeVarsAndDvarsMerged.Put("last_task_result", result)
					}
				}
			}

			func() {
				result := cache.StepStack.GetTop().(*cache.StepRuntimeContext).Result

				if result != nil && result.Code == 0 {
					u.LogOk(".")
				}

				if !u.Contains(step.Flags, "ignore_error") {
					if result != nil && result.Code != 0 {
						u.InvalidAndExit("Failed And Not Ignored!", "You may want to continue and ignore the error")
					}
				}
				if u.Contains(step.Flags, "pause") {
					u.Ppromptvvvvv("pause action to continue", " q: quit\n c: continue")
					reader := bufio.NewReader(os.Stdin)
					keyinput, _ := reader.ReadString('\n')

					switch keyinput {
					case "q\n":
						u.GraceExit("puase action", "client choce to stop continuing the execution")
					case "c\n":
						//contine
					default:
						//continue

					}

				}

			}()

			cache.StepStack.Pop()
		}

		execStep()

	}

}


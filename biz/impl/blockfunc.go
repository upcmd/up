// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/imdario/mergo"
	ms "github.com/mitchellh/mapstructure"
	"github.com/mohae/deepcopy"
	"github.com/stephencheng/up/biz"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
	ee "github.com/stephencheng/up/utils/error"
	"reflect"
	"strconv"
)

type BlockFuncAction struct {
	Do        interface{}
	Vars      *core.Cache
	Tasknames []string
	Steps     *Steps
}

func (f *BlockFuncAction) Adapt() {
	var flowname string
	var flow Steps

	switch f.Do.(type) {
	case string:
		//a flow name + refdir to load the flow
		raw := f.Do.(string)
		flowname = Render(raw, f.Vars)
		u.P(flowname)

	case []interface{}:
		//detailed steps
		err := ms.Decode(f.Do, &flow)
		u.LogErrorAndExit("load steps", err, "steps has configuration problem, please fix it")

	default:
		u.LogWarn("Block func", "Not implemented or void for no action!")
	}

	f.Steps = &flow
}

func (f *BlockFuncAction) Exec() {
	BlockFlowRun(f.Steps, f.Vars)
}

func BlockFlowRun(flow *Steps, execVars *core.Cache) {
	rtContext := BlockRuntimeContext{
		BlockBaseVars: execVars,
	}
	BlockStack().Push(&rtContext)

	//switch to test code
	//flow.ExecFlow()
	flow.Exec(true)
	BlockStack().Pop()
}

func (flow *Steps) ExecFlow() {

	for idx, step := range *flow {

		taskLayerCnt := TaskerRuntime().Tasker.TaskStack.GetLen()
		u.LogDesc("block step", idx+1, taskLayerCnt, step.Name, step.Desc)
		u.Ppmsgvvvv(step)

		execStep := func() {
			rtContext := StepRuntimeContext{
				Stepname: step.Name,
			}
			StepStack().Push(&rtContext)

			step.ExecTest()

			result := StepRuntime().Result
			taskname := TaskerRuntime().Tasker.TaskStack.GetTop().(*TaskRuntimeContext).Taskname

			//TODO: add support for block
			if u.Contains([]string{FUNC_SHELL, FUNC_CALL}, step.Func) {
				if step.Reg == "auto" {
					TaskRuntime().ExecbaseVars.Put(u.Spf("register_%s_%s", taskname, step.Name), result.Output)
				} else if step.Reg != "" {
					TaskRuntime().ExecbaseVars.Put(u.Spf("%s", step.Reg), result.Output)
				} else {
					if step.Func == FUNC_SHELL {
						TaskRuntime().ExecbaseVars.Put("last_result", result)
					}
				}
			}

			func() {
				result := StepRuntime().Result

				if result != nil && result.Code == 0 {
					u.LogOk(".")
				}

				if !u.Contains(step.Flags, "ignore_error") {
					if result != nil && result.Code != 0 {
						u.InvalidAndExit("Failed And Not Ignored!", "You may want to continue and ignore the error")
					}
				}

			}()

			StepStack().Pop()
		}

		if !TaskerRuntime().Tasker.TaskBreak {
			execStep()
		} else {
			TaskerRuntime().Tasker.TaskBreak = false
			break
		}

	}

}

func (step *Step) ExecTest() {
	var action biz.Do

	var bizErr *ee.Error = ee.New()
	var stepExecVars *core.Cache
	stepExecVars = step.getRuntimeExecVarsTest("get plain exec vars")
	//u.Ptmpdebug("99", stepExecVars)
	validation(stepExecVars)

	if step.Flags != nil && u.Contains(step.Flags, "pause") {
		pause(stepExecVars)
	}

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
			funcAction := CallFuncAction{
				Do:   step.Do,
				Vars: stepExecVars,
			}
			action = biz.Do(&funcAction)

		case FUNC_BLOCK:
			funcAction := BlockFuncAction{
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

		case "":
			u.InvalidAndExit("Step dispatch", "func name is empty and not defined")
			bizErr.Mark = "func name not implemented"

		default:
			u.InvalidAndExit("Step dispatch", u.Spf("func name(%s) is not recognised and implemented", step.Func))
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
						rawUtil := step.Until
						func() {
							//loop points to a var name which is a slice
							if reflect.TypeOf(step.Loop).Kind() == reflect.String {
								loopVarName := Render(step.Loop.(string), stepExecVars)
								loopObj := stepExecVars.Get(loopVarName)
								if loopObj == nil {
									u.InvalidAndExit("Evaluating loop var and object", u.Spf("Please use a correct varname:(%s) containing a list of values", loopVarName))
								}
								if reflect.TypeOf(loopObj).Kind() == reflect.Slice {
									switch loopObj.(type) {
									case []interface{}:
										for idx, item := range loopObj.([]interface{}) {
											routeFuncType(&LoopItem{idx, idx + 1, item})
											if rawUtil != "" {
												untilEval := Render(rawUtil, stepExecVars)
												toBreak, err := strconv.ParseBool(untilEval)
												u.LogErrorAndExit("evaluate until condition", err, u.Spf("please fix until condition evaluation: [%s]", untilEval))
												if toBreak {
													u.Pvvvv("loop util conditional break")
													break
												} else {
													chainAction(&action)
												}
											} else {
												chainAction(&action)
											}
										}

									case []string:
										for idx, item := range loopObj.([]string) {
											routeFuncType(&LoopItem{idx, idx + 1, item})
											if rawUtil != "" {
												untilEval := Render(rawUtil, stepExecVars)
												toBreak, err := strconv.ParseBool(untilEval)
												u.LogErrorAndExit("evaluate until condition", err, u.Spf("please fix until condition evaluation: [%s]", untilEval))
												if toBreak {
													u.Pvvvv("loop util conditional break")
													break
												} else {
													chainAction(&action)
												}
											} else {
												chainAction(&action)
											}
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
									if rawUtil != "" {
										untilEval := Render(rawUtil, stepExecVars)
										toBreak, err := strconv.ParseBool(untilEval)
										u.LogErrorAndExit("evaluate until condition", err, u.Spf("please fix until condition evaluation: [%s]", untilEval))
										if toBreak {
											u.Pvvvv("loop util conditional break")
											break
										} else {
											chainAction(&action)
										}
									} else {
										chainAction(&action)
									}
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
			IfEval := Render(step.If, stepExecVars)
			if IfEval != "<no value>" {
				goahead, err := strconv.ParseBool(IfEval)
				u.LogErrorAndExit("evaluate condition", err, u.Spf("please fix if condition evaluation: [%s]", IfEval))
				if goahead {
					dryRunOrContinue()
				} else {
					if step.Else != nil && step.Else != "" {
						doElse(step.Else, stepExecVars)
					} else {
						u.Pvvv("condition failed, skip executing step", step.Name)
					}
				}
			} else {
				u.Pvvv("condition failed, skip executing step", step.Name)
			}
		} else {
			dryRunOrContinue()
		}

	}()
}

func (step *Step) getRuntimeExecVarsTest(mark string) *core.Cache {
	var execvars core.Cache
	var resultVars *core.Cache

	execvars = deepcopy.Copy(*TaskRuntime().ExecbaseVars).(core.Cache)

	taskVars := TaskRuntime().TaskVars
	mergo.Merge(&execvars, taskVars, mergo.WithOverride)
	//u.Ptmpdebug("33", execvars)
	//u.Ptmpdebug("44", step.Vars)
	//if IsCalled() {
	//	u.Ptmpdebug("if", "if")
	//	if step.Vars != nil {
	//		mergo.Merge(&step.Vars, &execvars, mergo.WithOverride)
	//		resultVars = &step.Vars
	//	} else {
	//		resultVars = &execvars
	//	}
	//} else {
	//	u.Ptmpdebug("else", "else")
	//	mergo.Merge(&execvars, &step.Vars, mergo.WithOverride)
	//	resultVars = &execvars
	//}
	blockvars := BlockStack().GetTop().(*BlockRuntimeContext).BlockBaseVars
	mergo.Merge(&execvars, blockvars, mergo.WithOverride)
	mergo.Merge(&execvars, &step.Vars, mergo.WithOverride)
	resultVars = &execvars

	u.Pfvvvv("current exec runtime[%s] vars:", mark)
	u.Ppmsgvvvv(resultVars)
	//u.Ptmpdebug("55", resultVars)

	//so far the execvars includes: scope vars + scope dvars + global runtime vars + task vars
	resultVars = VarsMergedWithDvars("local", resultVars, &step.Dvars, resultVars)

	//so far the resultVars includes: the local vars + dvars rendered using execvars
	u.Ppmsgvvvhint("overall final exec vars:", resultVars)
	//u.Ptmpdebug("99", resultVars)
	return resultVars
}



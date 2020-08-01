// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/fatih/color"
	"github.com/imdario/mergo"
	ms "github.com/mitchellh/mapstructure"
	"github.com/mohae/deepcopy"
	"github.com/upcmd/up/biz"
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
	ee "github.com/upcmd/up/utils/error"
	"github.com/xlab/treeprint"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

type Step struct {
	Name     string
	Do       interface{} //FuncImpl
	Dox      interface{}
	Func     string
	Vars     core.Cache
	Dvars    Dvars
	Desc     string
	Reg      string
	Flags    []string
	If       string
	Else     interface{}
	Loop     interface{}
	Until    string
	RefDir   string
	VarsFile string
	Timeout  int //milli seconds, only for shell func
	Finally  interface{}
	Rescue   bool
}

type Steps []Step

/*
ExecbaseVars is the the scope containing passed in caller vars
local vars will be merged depending if it is a callee task
merge taskvars before final merge as task vars will be the calculated result of current stack
*/
func (step *Step) getRuntimeExecVars(fromBlock bool) *core.Cache {
	var execvars core.Cache
	var resultVars *core.Cache

	if u.Contains(step.Flags, "pure") {
		execvars = *core.NewCache()
	} else {
		execvars = deepcopy.Copy(*TaskRuntime().ExecbaseVars).(core.Cache)
	}

	mergo.Merge(&execvars, UpRunTimeVars, mergo.WithOverride)

	taskVars := TaskRuntime().TaskVars
	mergo.Merge(&execvars, taskVars, mergo.WithOverride)

	if step.VarsFile != "" {
		refdir := ConfigRuntime().RefDir
		if step.RefDir != "" {
			raw := step.RefDir
			refdir = Render(raw, execvars)
		}

		var varsfile string
		if step.VarsFile != "" {
			raw := step.VarsFile
			varsfile = Render(raw, execvars)
		}

		filepath := path.Join(refdir, varsfile)
		if _, err := os.Stat(filepath); !os.IsNotExist(err) {
			yamlvarsroot := u.YamlLoader("varsfile", refdir, varsfile)
			filevars := loadRefVars(yamlvarsroot)
			mergo.Merge(filevars, &step.Vars, mergo.WithOverride)
			step.Vars = *filevars
		} else {
			u.LogWarn("varsfile is not loaded, ignored", u.Spf("%s does not exist", filepath))
		}
	}

	if fromBlock {
		blockvars := BlockStack().GetTop().(*BlockRuntimeContext).BlockBaseVars
		mergo.Merge(&execvars, blockvars, mergo.WithOverride)
		mergo.Merge(&execvars, &step.Vars, mergo.WithOverride)
		resultVars = &execvars
	} else {
		if IsCalledTask() {
			//u.Ptmpdebug("if", "if")
			if step.Vars != nil {
				mergo.Merge(&step.Vars, &execvars, mergo.WithOverride)
				resultVars = &step.Vars
			} else {
				resultVars = &execvars
			}
		} else {
			//u.Ptmpdebug("else", "else")
			mergo.Merge(&execvars, &step.Vars, mergo.WithOverride)
			resultVars = &execvars
		}
	}
	u.Pfvvvv("current exec runtime vars:")
	u.Ppmsgvvvv(resultVars)

	StepRuntime().ContextVars = resultVars
	//so far the execvars includes: scope vars + scope dvars + global runtime vars + task vars
	varsWithDvars := VarsMergedWithDvars("local", resultVars, &step.Dvars, resultVars)

	//the processed varsWithDvars must merge with result vars: from varsWithDvars to resultVars
	//care taken to register new vars to dvar processing phase
	mergo.Merge(resultVars, varsWithDvars, mergo.WithOverride)

	//so far the resultVars includes: the local vars + dvars rendered using execvars
	u.Ppmsgvvvhint(u.Spf("%s: final context exec vars:", ConfigRuntime().ModuleName), resultVars)
	//debugVars()
	return resultVars
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

func validation(vars *core.Cache) {
	identified := false

	for k, _ := range *vars {
		if k == "" {
			u.InvalidAndPanic("validating var name", "var name can not be empty")
		}
		if u.CharIsNum(k[0:1]) != -1 {
			identified = true
			u.InvalidAndPanic("validating var name", u.Spf("var name (%s) can not start with number", k))
		}
	}

	if identified {
		u.InvalidAndPanic("vars validation", "please fix all validation before continue")
	}

}

func (step *Step) Exec(fromBlock bool) {
	var action biz.Do

	defer func() {
		if step.Finally != nil && step.Finally != "" {
			u.PlnBlue("Step Finally:")
			u.Ppmsg(StepRuntime().Result)
			u.Ppmsg(step.Vars)
		}
		paniced := false
		if step.Vars == nil {
			step.Vars = *core.NewCache()
		}

		step.Vars.Put(UP_RUNTIME_SHELL_EXEC_RESULT, StepRuntime().Result)
		//debugVars()
		var panicInfo interface{}
		if r := recover(); r != nil {
			u.PlnBlue(u.Spf("Recovered from: %s", r))
			paniced = true
			panicInfo = r
		}

		if step.Finally != nil && step.Finally != "" {
			execFinally(step.Finally, &step.Vars)
		}

		step.Vars.Delete(UP_RUNTIME_SHELL_EXEC_RESULT)

		if paniced && step.Rescue == false {
			u.LogWarn("No rescued in step level", "please assess the panic problem and cause, fix it before re-run the task")
			panic(panicInfo)
		} else if paniced {
			u.LogWarn("Rescued in step level, but not advised!", "setting rescue to yes/true to continue is not recommended\nit is advised to locate root cause of the problem, fix it and re-run the task again\nit is the best practice to test the execution in your ci pipeline to eliminate problems rather than dynamically fix using rescue")
		}
	}()

	var bizErr *ee.Error = ee.New()
	var stepExecVars *core.Cache

	stepExecVars = step.getRuntimeExecVars(fromBlock)
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
			u.InvalidAndPanic("Step dispatch", "func name is empty and not defined")
			bizErr.Mark = "func name not implemented"

		default:
			u.InvalidAndPanic("Step dispatch", u.Spf("func name(%s) is not recognised and implemented", step.Func))
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
									u.InvalidAndPanic("Evaluating loop var and object", u.Spf("Please use a correct varname:(%s) containing a list of values", loopVarName))
								}
								if reflect.TypeOf(loopObj).Kind() == reflect.Slice {
									switch loopObj.(type) {
									case []interface{}:
										for idx, item := range loopObj.([]interface{}) {
											routeFuncType(&LoopItem{idx, idx + 1, item})
											if rawUtil != "" {
												untilEval := Render(rawUtil, stepExecVars)
												toBreak, err := strconv.ParseBool(untilEval)
												u.LogErrorAndPanic("evaluate until condition", err, u.Spf("please fix until condition evaluation: [%s]", untilEval))
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
												u.LogErrorAndPanic("evaluate until condition", err, u.Spf("please fix until condition evaluation: [%s]", untilEval))
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

									case []int64:
										for idx, item := range loopObj.([]int64) {
											routeFuncType(&LoopItem{idx, idx + 1, item})
											if rawUtil != "" {
												untilEval := Render(rawUtil, stepExecVars)
												toBreak, err := strconv.ParseBool(untilEval)
												u.LogErrorAndPanic("evaluate until condition", err, u.Spf("please fix until condition evaluation: [%s]", untilEval))
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
									u.InvalidAndPanic("evaluate loop var", "loop var is not a array/list/slice")
								}
							} else if reflect.TypeOf(step.Loop).Kind() == reflect.Slice {
								//loop itself is a slice
								for idx, item := range step.Loop.([]interface{}) {
									routeFuncType(&LoopItem{idx, idx + 1, item})
									if rawUtil != "" {
										untilEval := Render(rawUtil, stepExecVars)
										toBreak, err := strconv.ParseBool(untilEval)
										u.LogErrorAndPanic("evaluate until condition", err, u.Spf("please fix until condition evaluation: [%s]", untilEval))
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
								u.InvalidAndPanic("evaluate loop items", "please either use a list or a template evaluation which could result in a value of a list")
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
		if step.If != NONE_VALUE && step.If != "" {
			IfEval := Render(step.If, stepExecVars)
			if IfEval != NONE_VALUE {
				goahead, err := strconv.ParseBool(IfEval)
				u.LogErrorAndPanic("evaluate condition", err, u.Spf("please fix if condition evaluation: [%s]", IfEval))
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

func doElse(elseCalls interface{}, execVars *core.Cache) {
	var taskname string
	var tasknames []string
	var flow Steps

	switch elseCalls.(type) {
	case string:
		taskname = elseCalls.(string)
		tasknames = append(tasknames, taskname)

	case []interface{}:
		elseStr := u.Spf("%s", elseCalls)
		if strings.Index(elseStr, "map") != -1 && strings.Index(elseStr, "func:") != -1 {
			err := ms.Decode(elseCalls, &flow)
			u.LogErrorAndPanic("load steps in else", err, "steps has configuration problem, please fix it")
			BlockFlowRun(&flow, execVars)
		} else {
			err := ms.Decode(elseCalls, &tasknames)
			u.LogErrorAndPanic("call func alias: else", err, "please ref to a task name only")
		}

	default:
		u.LogWarn("else ..", "Not implemented or void for no action!")
	}

	if len(tasknames) > 0 {
		for _, tmptaskname := range tasknames {
			taskname := Render(tmptaskname, execVars)
			u.PpmsgvvvvvhintHigh(u.Spf("else caller vars to task (%s):", taskname), execVars)
			ExecTask(taskname, execVars)
		}
	}

}

func execFinally(finally interface{}, execVars *core.Cache) {
	var taskname string
	var tasknames []string
	var flow Steps
	switch finally.(type) {
	case string:
		taskname = finally.(string)
		tasknames = append(tasknames, taskname)

	case []interface{}:
		elseStr := u.Spf("%s", finally)
		if strings.Index(elseStr, "map") != -1 && strings.Index(elseStr, "func:") != -1 {
			err := ms.Decode(finally, &flow)
			u.LogErrorAndPanic("load steps in finally", err, "steps/flow has configuration problem, please fix it")
			BlockFlowRun(&flow, execVars)
		} else {
			err := ms.Decode(finally, &tasknames)
			u.LogErrorAndPanic("load task names in finally", err, "please ref to a task name only")
		}

	default:
		u.LogWarn("finally ..", "Not implemented or void for no action!")
	}

	if len(tasknames) > 0 {
		for _, tmptaskname := range tasknames {
			taskname := Render(tmptaskname, execVars)
			u.PpmsgvvvvvhintHigh(u.Spf("finally caller vars to task (%s):", taskname), execVars)
			ExecTask(taskname, execVars)
		}
	}

}

func (steps *Steps) InspectSteps(tree treeprint.Tree, level *int) bool {
	for _, step := range *steps {
		desc := strings.Split(step.Desc, "\n")[0]
		if step.Func == FUNC_CALL {
			branch := tree.AddMetaBranch(func() string {
				if step.Loop != "" {
					return step.Name + color.HiYellowString("%s", " /call.")
				} else {
					return step.Name
				}
			}(), desc)
			var callee string
			switch t := step.Do.(type) {
			case string:
				callee = step.Do.(string)
				if !TaskerRuntime().Tasker.InspectTask(callee, branch, level) {
					break
				}
				*level -= 1
				//branch.AddBranch("aa")
			case []interface{}:
				calleeTasknames := step.Do.([]interface{})
				breakFlag := false
				for _, x := range calleeTasknames {
					callee = x.(string)
					if !TaskerRuntime().Tasker.InspectTask(callee, branch, level) {
						breakFlag = true
						break
					}
					*level -= 1
				}
				if breakFlag {
					break
				}
			default:
				u.Pf("type: %T", t)
			}

		} else if step.Func == FUNC_BLOCK {
			branch := tree.AddMetaBranch(func() string {
				if step.Loop != "" {
					return step.Name + color.HiYellowString("%s", " /block.")
				} else {
					return step.Name
				}
			}(), desc)

			switch t := step.Do.(type) {
			case string:
				rawFlowname := step.Do.(string)
				tree.AddNode(u.Spf("%s %s", color.HiYellowString("%s", " ..flow ->"), rawFlowname))

			case []interface{}:
				//detailed steps
				var steps Steps
				err := ms.Decode(step.Do, &steps)
				u.LogErrorAndPanic("load steps", err, "configuration problem, please fix it")
				steps.InspectSteps(branch, level)

			default:
				u.Pf("type: %T", t)
			}

		} else {
			tree.AddNode(u.Spf("%s: %s", step.Name, desc))
		}
	}
	return true
}

func (steps *Steps) Exec(fromBlock bool) {

	for idx, step := range *steps {

		taskLayerCnt := TaskerRuntime().Tasker.TaskStack.GetLen()
		u.LogDesc("step", idx+1, taskLayerCnt, step.Name, step.Desc)
		u.Ppmsgvvvvv(step)

		execStep := func() {
			rtContext := StepRuntimeContext{
				Stepname: step.Name,
				Timeout:  step.Timeout,
			}
			StepStack().Push(&rtContext)

			if step.Do == nil && step.Dox != nil {
				u.LogWarn("*", "Step is deactivated!")
			} else {
				step.Exec(fromBlock)
			}

			result := StepRuntime().Result
			taskname := TaskerRuntime().Tasker.TaskStack.GetTop().(*TaskRuntimeContext).Taskname

			//TODO: add support for block
			if u.Contains([]string{FUNC_SHELL, FUNC_CALL}, step.Func) {
				if step.Reg == "auto" {
					if step.Name == "" {
						TaskRuntime().ExecbaseVars.Put(u.Spf("%s_%d_result", taskname, idx), result)
					} else {
						TaskRuntime().ExecbaseVars.Put(u.Spf("%s_%s_result", taskname, step.Name), result)
					}
				} else if step.Reg != "" {
					TaskRuntime().ExecbaseVars.Put(u.Spf("%s", step.Reg), result)
				}
				if step.Func == FUNC_SHELL {
					TaskRuntime().ExecbaseVars.Put("last_result", result)
				}

			}

			func() {
				result := StepRuntime().Result

				if result != nil && result.Code == 0 {
					u.LogOk(".")
				}

				if !u.Contains(step.Flags, "ignoreError") {
					if result != nil && result.Code != 0 {
						u.InvalidAndPanic("Failed And Not Ignored!", "You may want to continue and ignore the error")
					}
				} else {
					if result != nil && result.Code != 0 {
						u.LogWarn("HightLight:", "Error ignored!!!")
					}
				}

			}()

			StepStack().Pop()
		}

		if !TaskerRuntime().Tasker.TaskBreak {
			execStep()
		} else {
			TaskerRuntime().Tasker.TaskBreak = false
			u.LogWarn("break", "client chose to break")
			break
		}

	}

}

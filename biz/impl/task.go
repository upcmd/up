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
	"github.com/spf13/viper"
	"github.com/stephencheng/up/model"
	"github.com/stephencheng/up/model/core"
	"github.com/stephencheng/up/model/stack"
	u "github.com/stephencheng/up/utils"
	"github.com/xlab/treeprint"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func InitDefaultSkeleton() {
	filepath := path.Join(".", "upconfig.yml")
	ioutil.WriteFile(filepath, []byte(u.DEFAULT_CONFIG), 0644)
	filepath = path.Join(".", "up.yml")
	ioutil.WriteFile(filepath, []byte(u.DEFAULT_UP_TASK_YML), 0644)
}

type Tasker struct {
	TaskYmlRoot  *viper.Viper
	Tasks        *model.Tasks
	InstanceName string
	Dryrun       bool
	TaskStack    *stack.ExecStack
	StepStack    *stack.ExecStack
	BlockStack   *stack.ExecStack
	TaskBreak    bool
}

type TaskerRuntimeContext struct {
	Taskername string
	Tasker     *Tasker
	//TaskVars        *Cache
	//ReturnVars      *Cache
}

func NewTasker(id string) *Tasker {
	priorityLoadingTaskFile := filepath.Join(".", u.CoreConfig.TaskFile)
	refDir := "."
	if _, err := os.Stat(priorityLoadingTaskFile); err != nil {
		refDir = u.CoreConfig.RefDir
	}

	taskYmlRoot := u.YamlLoader("Task", refDir, u.CoreConfig.TaskFile)
	tasker := &Tasker{
		TaskYmlRoot: taskYmlRoot,
	}

	tasker.initRuntime()

	taskerContext := TaskerRuntimeContext{
		Taskername: "abcde",
		Tasker:     tasker,
		//TODO: use namegen to generate random name in config default settings
	}

	TaskerStack.Push(&taskerContext)
	tasker.SetInstanceName(id)
	//TODO: refactory of the runtime init after config is loaded to a proper place
	FuncMapInit()
	tasker.loadScopes()
	ScopeProfiles.InitContextInstances()
	tasker.loadRuntimeGlobalVars()
	tasker.loadRuntimeDvars()
	SetRuntimeVarsMerged(tasker.InstanceName)
	SetRuntimeGlobalMergedWithDvars()
	tasker.loadTasks()

	return tasker
}

func (t *Tasker) SetInstanceName(id string) {
	if id != "" {
		t.InstanceName = id
	} else {
		t.InstanceName = "nonamed"
	}
}

func (t *Tasker) initRuntime() {
	//InstanceName string
	//Dryrun       bool
	t.TaskStack = stack.New("task")
	t.StepStack = stack.New("step")
	t.BlockStack = stack.New("block")
	//TaskBreak    bool

}

func (t *Tasker) ListTasks() {
	caps := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	u.Pln("-task list")
	maxlen := 0
	for _, task := range *t.Tasks {
		tasknamelen := len(task.Name)
		if tasknamelen > maxlen {
			maxlen = tasknamelen
		}
	}
	format := "  %4d  | %" + u.Spf("%d", maxlen) + "s: |%9s| %s "
	for idx, task := range *t.Tasks {
		start := task.Name[0:1]
		if strings.Contains(caps, start) {
			color.HiGreen("%s", u.Spf(format, idx+1, task.Name, "public", task.Desc))
		} else {
			color.Yellow("%s", u.Spf(format, idx+1, task.Name, "protected", task.Desc))
		}

		u.Ppmsgvvvv(task)
	}
	u.Pln("-\n")
}

func (t *Tasker) ListAllTasks() {
	u.Pln("-inspect all tasks:")
	for _, task := range *t.Tasks {
		t.ListTask(task.Name)
	}
}

func (tasker *Tasker) ListTask(taskname string) {
	var tree = treeprint.New()
	//u.Pln("\ninspect task:")
	level := 0
	for _, task := range *tasker.Tasks {
		if task.Name == taskname {
			desc := strings.Split(task.Desc, "\n")[0]
			u.Pf("%s: %s", color.BlueString("%s", task.Name), desc)
			var steps Steps
			err := ms.Decode(task.Task, &steps)
			u.LogErrorAndExit("decode steps:", err, "please fix data type in yaml config")

			for _, step := range steps {
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
						if !tasker.InspectTask(callee, branch, &level) {
							break
						}
						level -= 1
						//branch.AddBranch("aa")
					case []interface{}:
						calleeTasknames := step.Do.([]interface{})
						breakFlag := false
						for _, x := range calleeTasknames {
							callee = x.(string)
							if !tasker.InspectTask(callee, branch, &level) {
								breakFlag = true
								break
							}
							level -= 1
						}
						if breakFlag {
							break
						}
					default:
						u.Pf("type: %T", t)
					}

				} else {
					tree.AddNode(u.Spf("%s: %s", step.Name, desc))
				}
			}
		}
	}
	u.Pln(tree.String())
}

func (tasker *Tasker) InspectTask(taskname string, branch treeprint.Tree, level *int) bool {
	*level += 1
	maxLayers, _ := strconv.Atoi(u.CoreConfig.MaxCallLayers)
	if *level > maxLayers {
		u.LogWarn("evaluate max task stack layer", "please setup max MaxCallLayers correctly, or fix recursive cycle calls")
		return false
	}
	for _, task := range *tasker.Tasks {
		if task.Name == taskname {
			desc := strings.Split(task.Desc, "\n")[0]
			br := branch.AddMetaBranch(color.BlueString("%s", task.Name), desc)
			var steps Steps
			err := ms.Decode(task.Task, &steps)
			u.LogErrorAndExit("decode steps:", err, "please fix data type in yaml config")

			for _, step := range steps {
				desc := strings.Split(step.Desc, "\n")[0]
				if step.Func == FUNC_CALL {
					var callee string
					switch t := step.Do.(type) {
					case string:

						brnode := br.AddMetaBranch(func() string {
							if step.Loop != nil {
								return step.Name + color.HiYellowString("%s", " /loop..")
							} else {
								return step.Name
							}
						}(), desc)

						callee = step.Do.(string)
						tasker.InspectTask(callee, brnode, level)
					case []interface{}:
						calleeTasknames := step.Do.([]interface{})
						for _, x := range calleeTasknames {
							brnode := br.AddMetaBranch(func() string {
								if step.Loop != "" {
									return step.Name + color.HiYellowString("%s", " /loop..")
								} else {
									return step.Name
								}
							}(), desc)

							callee = x.(string)
							tasker.InspectTask(callee, brnode, level)
						}
					default:
						u.Pf("type: %T", t)
					}
				} else {
					br.AddNode(u.Spf("%s: %s", step.Name, desc))
				}
			}
		}
	}
	return true
}

func (t *Tasker) ValidateTask(taskname string) {
	SetDryrun()
	t.ExecTask(taskname, nil)
}

func ExecTask(fulltaskname string, callerVars *core.Cache) {
	var modname string
	var taskname string

	func() {
		subnames := strings.Split(fulltaskname, ".")
		if len(subnames) > 2 {
			u.InvalidAndExit("task name validation", "task naming pattern: modulename.taskname")
		}

		if len(subnames) == 1 {
			modname = "self"
			taskname = subnames[0]
		} else if len(subnames) == 2 {
			modname = subnames[0]
			taskname = subnames[1]
		}
	}()

	if modname == "self" {
		TaskerRuntime().Tasker.ExecTask(taskname, callerVars)
	} else {
		//TODO: load the external module
		//change workdir to that dir and load task entry
		eTasker := NewTasker("something_to_be_defined")
		//TODO: to implemente
		eTasker.ExecTask(taskname, callerVars)
		TaskerStack.Pop()
	}

}

func (t *Tasker) ExecTask(taskname string, callerVars *core.Cache) {
	found := false
	for idx, task := range *t.Tasks {
		if taskname == task.Name {
			u.Pfvvvv("  located task-> %d [%s]: \n", idx+1, task.Name)

			var ctxCallerTaskname string
			if TaskerRuntime().Tasker.TaskStack.GetLen() > 0 {
				ctxCallerTaskname = TaskRuntime().TasknameLayered
			} else {
				ctxCallerTaskname = taskname
			}

			taskLayerCnt := TaskerRuntime().Tasker.TaskStack.GetLen()
			u.LogDesc("task", idx+1, taskLayerCnt, u.Spf("%s ==> %s", ctxCallerTaskname, taskname), task.Desc)
			found = true
			var steps Steps
			err := ms.Decode(task.Task, &steps)

			u.LogErrorAndExit("decode steps:", err, "please fix data type in yaml config")
			func() {
				//step name validation
				invalidNames := []string{}
				for _, step := range steps {
					if strings.Contains(step.Name, "-") {
						invalidNames = append(invalidNames, step.Name)
					}
				}

				if len(invalidNames) > 0 {
					u.InvalidAndExit(u.Spf("validating step name fails: %s ", invalidNames), "task name can not contain '-', please use '_' instead, failed names:")
				}
			}()

			func() {
				rtContext := TaskRuntimeContext{
					Taskname: taskname,
					TaskVars: core.NewCache(),
				}

				if IsAtRootTaskLevel() {
					rtContext.ExecbaseVars = RuntimeVarsAndDvarsMerged
					rtContext.TasknameLayered = taskname
				} else {
					rtContext.ExecbaseVars = callerVars
					rtContext.TasknameLayered = u.Spf("%s/%s", TaskRuntime().TasknameLayered, taskname)
				}
				rtContext.ExecbaseVars.Put(UP_RUNTIME_TASK_LAYER_NUMBER, TaskerRuntime().Tasker.TaskStack.GetLen())

				TaskerRuntime().Tasker.TaskStack.Push(&rtContext)
				u.Pvvvv("Executing task stack layer:", TaskerRuntime().Tasker.TaskStack.GetLen())
				maxLayers, err := strconv.Atoi(u.CoreConfig.MaxCallLayers)
				u.LogErrorAndExit("evaluate max task stack layer", err, "please setup max MaxCallLayers correctly")

				if maxLayers != 0 && TaskerRuntime().Tasker.TaskStack.GetLen() > maxLayers {
					u.LogError("Task exec stack layer check:", u.Spf("Too many layers of task executions, max allowed(%d), please fix your recursive call", maxLayers))
					os.Exit(-1)
				}

				steps.Exec(false)

				returnVars := TaskRuntime().ReturnVars
				TaskerRuntime().Tasker.TaskStack.Pop()
				if TaskerRuntime().Tasker.TaskStack.GetLen() > 0 && returnVars != nil {
					mergo.Merge(TaskRuntime().ExecbaseVars, returnVars, mergo.WithOverride)
				} else if TaskerRuntime().Tasker.TaskStack.GetLen() == 0 && returnVars != nil {
					mergo.Merge(RuntimeVarsAndDvarsMerged, returnVars, mergo.WithOverride)
				}
			}()

		}
	}

	if !found {
		u.Pferror("Task %s is not defined!", taskname)
		t.ListTasks()
	}
}

func (t *Tasker) validateAndLoadTaskRef() {
	//validation

	invalidNames := []string{}
	for idx, task := range *t.Tasks {
		if strings.Contains(task.Name, "-") {
			invalidNames = append(invalidNames, task.Name)
		}

		if task.Task != nil && task.Ref != "" {
			u.InvalidAndExit("validate task node and ref", "task and ref can not coexist")
		}

		//load ref task
		refdir := u.CoreConfig.RefDir

		if task.Ref != "" {
			if task.RefDir != "" {
				rawdir := task.RefDir
				refdir = Render(rawdir, RuntimeVarsAndDvarsMerged)
			}

			rawref := task.Ref
			ref := Render(rawref, RuntimeVarsAndDvarsMerged)

			yamlflowroot := u.YamlLoader("flow ref", refdir, ref)
			flow := loadRefFlow(yamlflowroot)
			(*t.Tasks)[idx].Task = flow
		}
	}

	if len(invalidNames) > 0 {
		u.InvalidAndExit(u.Spf("validating task name fails: %s ", invalidNames), "task name can not contain '-', please use '_' instead, failed names:")
	}
}

func (t *Tasker) loadRefTasks() {
	tasksRefList := t.TaskYmlRoot.Get("tasksref")
	if tasksRefList != nil {
		for _, ref := range tasksRefList.([]interface{}) {
			tasksYamlName := ref.(string)
			tasksYmlRoot := u.YamlLoader(tasksYamlName, u.CoreConfig.RefDir, tasksYamlName)

			var tasks model.Tasks
			tasksData := tasksYmlRoot.Get("tasks")
			err := ms.Decode(tasksData, &tasks)
			u.LogErrorAndExit(u.Spf("decode tasks:%s", tasksYamlName), err, "please fix configuration in tasks yaml file")
			for _, task := range tasks {
				*t.Tasks = append(*t.Tasks, task)
			}
		}
	}
}

func (t *Tasker) loadTasks() error {
	tasksData := t.TaskYmlRoot.Get("tasks")
	var tasks model.Tasks
	err := ms.Decode(tasksData, &tasks)
	u.LogErrorAndExit("decode tasks:main", err, "please fix configuration in tasks yaml file")
	t.Tasks = &tasks
	t.loadRefTasks()
	t.validateAndLoadTaskRef()

	return err
}

func loadRefFlow(yamlroot *viper.Viper) *Steps {
	flowData := yamlroot.Get("flow")
	var flow Steps
	err := ms.Decode(flowData, &flow)
	u.LogErrorAndExit("load ref flow", err, "flow of the steps has configuration problem, please fix it")
	return &flow
}

func (t *Tasker) loadScopes() {
	scopesData := t.TaskYmlRoot.Get("scopes")
	var scopes Scopes
	err := ms.Decode(scopesData, &scopes)
	SetScopeProfiles(&scopes)

	u.LogErrorAndExit("load full scopes", err, "please assess your scope configuration carefully")
}

func (t *Tasker) loadRuntimeGlobalVars() {
	varsData := t.TaskYmlRoot.Get("vars")
	var vars core.Cache
	err := ms.Decode(varsData, &vars)
	u.LogError("loadRuntimeGlobalVars", err)
	SetRuntimeGlobalVars(&vars)
}

func (t *Tasker) loadRuntimeDvars() *Dvars {
	dvarsData := t.TaskYmlRoot.Get("dvars")
	var dvars Dvars
	err := ms.Decode(dvarsData, &dvars)
	u.LogErrorAndExit("loadRuntimeDvars",
		err,
		"You must fix the data type to be\n string for a dvar value and try again. possible problems:\nthe name can not be single character 'y' or 'n' ",
	)

	//dvars.ValidateAndLoading()
	SetRuntimeGlobalDvars(&dvars)
	return &dvars
}


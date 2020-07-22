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
	"github.com/spf13/viper"
	"github.com/upcmd/up/model"
	"github.com/upcmd/up/model/core"
	"github.com/upcmd/up/model/stack"
	u "github.com/upcmd/up/utils"
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
	TaskYmlRoot      *viper.Viper
	Tasks            *model.Tasks
	InstanceName     string
	ExecProfilename  string
	Dryrun           bool
	TaskStack        *stack.ExecStack
	StepStack        *stack.ExecStack
	BlockStack       *stack.ExecStack
	TaskBreak        bool
	Config           *u.UpConfig
	GroupMembersList []string
	MemberGroupMap   map[string]string
	//expanded context only contains group and global scope, but not each instance vars
	ExpandedContext ScopeContext
	ScopeProfiles   *Scopes
	ExecProfiles    *ExecProfiles
	//this is the merged vars from within scope: global, groups level (if there is), instance varss, then global runtime vars
	RuntimeVarsMerged  *core.Cache
	ExecProfileEnvVars *core.Cache
	//this is the merged vars and dvars to a vars cache from within scope: global, groups level (if there is), instance varss, then global runtime vars
	//this vars should be used instead of RuntimeVarsMerged as it include both runtime vars and dvars except the local vars and dvars
	RuntimeVarsAndDvarsMerged *core.Cache
	RuntimeGlobalVars         *core.Cache
	RuntimeGlobalDvars        *Dvars
}

type ScopeContext map[string]*core.Cache
type ContextInstances []ScopeContext

type TaskerRuntimeContext struct {
	Tasker       *Tasker
	TaskerCaller *Tasker
	//TaskVars        *Cache
	//ReturnVars      *Cache
}

func NewTasker(instanceId string, eprofiename string, cfg *u.UpConfig) *Tasker {
	priorityLoadingTaskFile := filepath.Join(".", cfg.TaskFile)
	refDir := "."
	if _, err := os.Stat(priorityLoadingTaskFile); err != nil {
		refDir = cfg.RefDir
	}

	taskYmlRoot := u.YamlLoader("Task", refDir, cfg.TaskFile)
	tasker := &Tasker{
		TaskYmlRoot:      taskYmlRoot,
		MemberGroupMap:   map[string]string{},
		GroupMembersList: []string{},
		ExpandedContext:  ScopeContext{},
	}
	tasker.Config = cfg
	tasker.initRuntime()

	taskerContext := TaskerRuntimeContext{
		Tasker: tasker,
	}

	TaskerStack.Push(&taskerContext)

	tasker.loadExecProfiles()
	tasker.setInstanceName(instanceId, eprofiename)
	tasker.loadScopes()
	tasker.loadInstancesContext()
	tasker.loadRuntimeGlobalVars()
	tasker.loadRuntimeGlobalDvars()
	tasker.loadExecProfileEnvVars()
	tasker.MergeUptoRuntimeGlobalVars()
	tasker.MergeRuntimeGlobalDvars()
	tasker.loadTasks()

	return tasker
}

/*
Get the merged vars for specific scope instance
Validate the scopes
1. for the scope name equal to global, there should be no value for members, otherwise errors
2. for the scope with group members, it is a group itself
3. for the scope with no members and name is not global, it is a final instance
*/

func (t *Tasker) loadInstancesContext() {
	ss := t.ScopeProfiles
	//validation
	for idx, s := range *ss {

		if s.Ref != "" && s.Vars != nil {
			u.Dvvvvv(s)
			u.InvalidAndExit("verify scope ref and member coexistence", "ref and members can not both exist")
		}
		refdir := ConfigRuntime().RefDir
		if s.Ref != "" {
			if s.RefDir != "" {
				refdir = s.RefDir
			}
			yamlvarsroot := u.YamlLoader("ref vars", refdir, s.Ref)
			vars := *loadRefVars(yamlvarsroot)
			u.Pvvvv("loading vars from:", s.Ref)
			u.Ppmsgvvvv(vars)
			(*ss)[idx].Vars = vars
		}

	}

	u.Pvvvvv("-------full vars in scopes------")
	//u.Dpplnvvvv(ss)
	u.Dvvvvv(ss)

	var globalScope *Scope
	for idx, s := range *ss {
		if s.Name == "global" {
			if s.Members != nil {
				u.InvalidAndExit("scope expand", "global scope should not contains members")
			}
			globalScope = &(*ss)[idx]
		}
	}

	//expand dvars into global scope's vars space
	var globalvarsMergedWithDvars *core.Cache
	if globalScope != nil {
		globalvarsMergedWithDvars = GlobalVarsMergedWithDvars(globalScope)
	} else {
		globalvarsMergedWithDvars = core.NewCache()
	}

	for idx, s := range *ss {
		if s.Members != nil {
			for _, m := range s.Members {
				if u.Contains(t.GroupMembersList, m) {
					u.InvalidAndExit("scope expand", u.Spfv("duplicated member: %s\n", m))
				}
				t.GroupMembersList = append(t.GroupMembersList, m)
				t.MemberGroupMap[m] = s.Name
			}

			var groupvars core.Cache = deepcopy.Copy(*globalvarsMergedWithDvars).(core.Cache)
			mergo.Merge(&groupvars, s.Vars, mergo.WithOverride)

			//expand dvars into group scope's vars space
			groupScope := &(*ss)[idx]
			var groupvarsMergedWithDvars *core.Cache = ScopeVarsMergedWithDvars(groupScope, &groupvars)

			t.ExpandedContext[s.Name] = groupvarsMergedWithDvars
		}
	}

	t.ExpandedContext["global"] = globalvarsMergedWithDvars
	func() {
		u.Pvvvv("---------group vars----------")
		for k, v := range t.ExpandedContext {
			u.Pfvvvv("%s: %s", k, u.Sppmsg(*v))
		}
		u.Pfvvvv("groups members:%s\n", t.GroupMembersList)

	}()

}

func (t *Tasker) MergeRuntimeGlobalDvars() {
	var mergedVars core.Cache
	mergedVars = deepcopy.Copy(*t.RuntimeVarsMerged).(core.Cache)

	expandedVars := t.RuntimeGlobalDvars.Expand("runtime global", t.RuntimeVarsMerged)

	if t.RuntimeGlobalDvars != nil {
		mergo.Merge(&mergedVars, *expandedVars, mergo.WithOverride)
	}

	t.RuntimeVarsAndDvarsMerged = &mergedVars
	u.Ppmsgvvvvhint("-------runtime global final merged with dvars-------", mergedVars)
}

func (t *Tasker) loadExecProfileEnvVars() {
	var envVars *core.Cache = core.NewCache()
	var evars *EnvVars
	if p := t.getExecProfile(t.ExecProfilename); p != nil {

		if p.Ref != "" && p.Evars != nil {
			u.InvalidAndExit("exec proile validation", "You can only setup either ref file to load the env vars or use evars tag to config env vars, but not both")
		}

		refdir := ConfigRuntime().RefDir

		if p.Ref != "" {
			if p.RefDir != "" {
				refdir = p.RefDir
			}
			yamlevarsroot := u.YamlLoader("ref evars", refdir, p.Ref)
			evars = loadRefEvars(yamlevarsroot)
			u.Pvvvv("loading vars from:", p.Ref)
			u.Ppmsgvvvv(evars)
		}

		if p.Evars != nil {
			evars = &p.Evars
		}

		if evars != nil {
			for _, v := range *evars {
				envvarName := u.Spf("%s_%s", "envVar", v.Name)
				envVars.Put(envvarName, v.Value)
				os.Setenv(v.Name, v.Value)
			}
		}
	}

	t.ExecProfileEnvVars = envVars
	u.Ppmsgvvvhint(u.Spf("profile - %s envVars:", t.ExecProfilename), envVars)
}

//clear up everything in scope and cache
func (t *Tasker) Unset() {
	t.ExpandedContext = ScopeContext{}
	t.GroupMembersList = []string{}
	t.MemberGroupMap = map[string]string{}
	t.ScopeProfiles = nil
	t.RuntimeVarsMerged = nil
	t.RuntimeVarsAndDvarsMerged = nil
	t.RuntimeGlobalVars = nil
	t.RuntimeGlobalDvars = nil
	TaskerStack = stack.New("tasker")
}

/*
This will generate a one off vars merged from top level down to runtime
global and merge them all together,the result vars will be used to finally
merge with local func vars to be used in runtime execution time

pass in runtime id, if runtime id is in member list, eg dev -> nonprod
then merge runtimevars to group(nonprod)'s varss,

if runtime id (nonname) is not in member list,
then merge runtimevars to global varss,

This has chained dvar expansion through global to group then to instance level
and finally merge with global var, except the global dvars
*/
func (t *Tasker) MergeUptoRuntimeGlobalVars() {
	u.Pf("module: [%s], instance id: [%s], exec profile: [%s]\n", ConfigRuntime().ModuleName, t.InstanceName, t.ExecProfilename)
	var runtimevars core.Cache
	runtimevars = deepcopy.Copy(*t.ExpandedContext["global"]).(core.Cache)

	if u.Contains(t.GroupMembersList, t.InstanceName) {
		groupname := t.MemberGroupMap[t.InstanceName]
		//TODO: t.ExpandedContext[groupname] should have already merge to global vars, double check to confirm
		mergo.Merge(&runtimevars, *t.ExpandedContext[groupname], mergo.WithOverride)
		instanceVars := t.ScopeProfiles.GetInstanceVars(t.InstanceName)
		if instanceVars != nil {
			mergo.Merge(&runtimevars, instanceVars, mergo.WithOverride)
		}
	}

	//merge dvars for the instance
	//TODO: is this a duplication of: GetInstanceVars above?
	var instanceScope *Scope
	for idx, s := range *t.ScopeProfiles {
		if s.Name == t.InstanceName {
			instanceScope = &(*t.ScopeProfiles)[idx]
		}
	}

	var instanceVarsMergedWithDvars *core.Cache
	if instanceScope != nil {
		instanceVarsMergedWithDvars = VarsMergedWithDvars(instanceScope.Name, &instanceScope.Vars, &instanceScope.Dvars, &runtimevars)
		//merge back the expanded merged scope vars and dvars
		mergo.Merge(&runtimevars, *instanceVarsMergedWithDvars, mergo.WithOverride)
	}

	//merge with global vars
	mergo.Merge(&runtimevars, *t.RuntimeGlobalVars, mergo.WithOverride)

	u.Pfvvvv("merged[ %s ] runtime vars:", t.InstanceName)
	u.Ppmsgvvvv(runtimevars)
	u.Dvvvvv(runtimevars)

	t.RuntimeVarsMerged = &runtimevars
}

func (t *Tasker) setInstanceName(id, eprofilename string) {
	t.ExecProfilename = eprofilename
	instanceName := "nonamed"
	if id != "" {
		instanceName = id
	} else {
		if p := t.getExecProfile(eprofilename); p != nil {
			if p.Instance != "" {
				instanceName = p.Instance
			}
		}
	}

	t.InstanceName = instanceName
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

func (t *Tasker) LockModules() {
	if !t.ValidateAllModules() {
		u.InvalidAndExit("modules configuration is not valid", "please fix the problem and try again")
	}
	u.Pln("-lock repos:")

	lockMap := u.ModuleLockMap{}

	mlist := (*ConfigRuntime()).Modules
	if mlist != nil {
		for _, m := range mlist {
			m.Normalize()
			m.ShowDetails()
			gitdir := path.Join(m.Dir, ".git")
			if _, err := os.Stat(gitdir); !os.IsNotExist(err) {
				rev := u.GetHeadRev(m.Dir)
				lockMap[m.Alias] = rev
			}
		}
	}

	u.Pln("versions:")
	u.Ppmsg(lockMap)
	lockYml := core.ObjToYaml(lockMap)
	ioutil.WriteFile("./modlock.yml", []byte(lockYml), 0644)
	u.Pf("Please check in: [%s] into code repo", "modlock.yml")
}

func (t *Tasker) CleanModules() {

	if !t.ValidateAllModules() {
		u.InvalidAndExit("modules configuration is not valid", "please fix the problem and try again")
	}
	u.Pln("-clean repos:")
	//u.Pdebug(u.MainConfig.AbsWorkDir, u.GetDefaultModuleDir())
	//TODO
}

func (t *Tasker) PullModules() {
	if !t.ValidateAllModules() {
		u.InvalidAndExit("modules configuration is not valid", "please fix the problem and try again")
	}

	u.Pln("-pull repos:")

	mainMods := listModules("-main direct modules:", "%s/%s")
	clonedMainModNames := mainMods.PullMainModules()
	clonedSubModNames := append(clonedMainModNames, []string{}...)
	mainMods.PullCascadedModules(&clonedMainModNames, &clonedSubModNames)
}

func (t *Tasker) ValidateAllModules() bool {
	u.Pln("-validate all modules:")
	mlist := (*ConfigRuntime()).Modules

	namelist := []string{}
	policies := []string{"manual", "always", "skip"}
	errCnt := 0
	for idx, m := range mlist {
		m.Normalize()
		if u.Contains(namelist, m.Alias) {
			u.LogErrorMsg("alias duplication error", u.Spf("%d:%s", idx+1, m.Alias))
			errCnt += 1
		} else {
			namelist = append(namelist, m.Alias)
		}

		if m.Repo != "" && !u.Contains(policies, m.PullPolicy) {
			u.LogErrorMsg("pullpolicy error", u.Spf("%d:%s", idx+1, "must be one of: manual | always | skip"))
			errCnt += 1
		}

		if m.Repo != "" && m.Subdir != "" && m.Alias == "" {
			u.LogErrorMsg("alias must be set", u.Spf("%d:%s", idx+1, "alias is needed to avoid confusion"))
			errCnt += 1
		}
	}

	if errCnt == 0 {
		return true
	} else {
		return false
	}

}

//list tasker modules
func (t *Tasker) ListMainModules() {
	u.Pln("-list all modules:")
	mlist := ConfigRuntime().Modules
	mlist.ReportModules()
	t.ValidateAllModules()
}

//probing modules list all modules, including the main direct modules and the all indirect modules
func ListAllModules() {
	u.Pln("-list all modules:")
	mods := listModules("-main direct modules:", "%s/%s")
	u.Pln("- Insights:")
	mods.ReportModules()
	u.Pln("")
	mods = listModules("-indirect sub modules:", "%s/.upmodules/*/%s")
	u.Pln("- Insights:")
	mods.ReportModules()
}

func listModules(desc, pattern string) *u.Modules {
	cfgname := "upconfig.yml"
	filelist := []string{}
	match := u.Spfv(pattern, u.MainConfig.AbsWorkDir, cfgname)
	files, err := filepath.Glob(match)
	u.LogError("list upconfig.yml", err)
	filelist = append(filelist, files...)

	modlist := u.Modules{}
	for _, f := range filelist {
		cfg := u.NewUpConfig(path.Dir(f), cfgname)
		modlist = append(modlist, cfg.Modules...)
	}
	u.Pf("\n%s\n", desc)
	yml := core.ObjToYaml(modlist)
	u.Pln(yml)

	return &modlist
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
			steps.InspectSteps(tree, &level)
		}
	}
	u.Pln(tree.String())
}

func (tasker *Tasker) InspectTask(taskname string, branch treeprint.Tree, level *int) bool {
	*level += 1
	maxLayers, _ := strconv.Atoi(ConfigRuntime().MaxCallLayers)
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
				} else if step.Func == FUNC_BLOCK {
					branch := branch.AddMetaBranch(func() string {
						if step.Loop != "" {
							return step.Name + color.HiYellowString("%s", " /block.")
						} else {
							return step.Name
						}
					}(), desc)

					switch t := step.Do.(type) {
					case string:
						rawFlowname := step.Do.(string)
						branch.AddNode(u.Spf("%s %s", color.HiYellowString("%s", " ..flow ->"), rawFlowname))

					case []interface{}:
						//detailed steps
						var steps Steps
						err := ms.Decode(step.Do, &steps)
						u.LogErrorAndExit("load steps", err, "configuration problem, please fix it")
						steps.InspectSteps(branch, level)

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
	t.ExecTask(taskname, nil, false)
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
		TaskerRuntime().Tasker.ExecTask(taskname, callerVars, false)
	} else {
		if modname == GetBaseModuleName() {
			u.InvalidAndExit("module name should not be the same as the main caller", "please check your task configuration")
		} else {

			cwd, err := os.Getwd()

			if err != nil {
				u.LogErrorAndExit("cwd", err, "working directory error")
			}

			mods := TaskerRuntime().Tasker.Config.Modules
			//u.Pdebug(TaskerRuntime().Tasker.Config)
			if mods != nil {
				mod := TaskerRuntime().Tasker.Config.Modules.LocateModule(modname)
				//u.Pdebug(cwd, mod)
				//mdir := "hello-module/"
				//iid := "dev"

				if mod != nil {
					func() {
						//TODO: exclude the subdir case
						var modpath string
						if path.IsAbs(mod.Dir) {
							modpath = mod.Dir
						} else {
							modpath = path.Clean(path.Join(BaseDir, mod.Dir))
						}
						os.Chdir(modpath)
						if _, err := os.Stat(modpath); !os.IsNotExist(err) {
							/*
								in module loading, since you can not pass in the cli options, so:
								version: will not be used at all
								Verbose: determined by caller, so not relevant
								MaxCallLayers: determined by caller
								RefDir: applied
								TaskFile: applied
								ConfigDir: will not be used at all since no cli option to override this, it will be always be current dir .
								ConfigFile: will not be used at all since no cli option to override this, it will be always be upconfig.yml from default
							*/
							mcfg := u.NewUpConfig("", "")
							mcfg.SetModulename(modname)
							mcfg.InitConfig()
							taskerCaller := TaskerRuntime().Tasker
							mTasker := NewTasker(mod.Iid, "", mcfg)
							TaskerRuntime().TaskerCaller = taskerCaller
							u.Pf("=>call module: [%s] task: [%s]\n", modname, taskname)
							//u.Ptmpdebug("55", callerVars)

							func() {
								taskerLayer := TaskerStack.GetLen()
								UpRunTimeVars.Put(UP_RUNTIME_TASKER_LAYER_NUMBER, taskerLayer)
								u.Pvvvv("Executing tasker layer:", taskerLayer)
								maxLayers, err := strconv.Atoi(u.MainConfig.MaxModuelCallLayers)
								u.LogErrorAndExit("evaluate max tasker module call layer", err, "please setup max MaxModuelCallLayers properly for your case")

								if maxLayers != 0 && taskerLayer > maxLayers {
									u.InvalidAndExit("Module call layer check:", u.Spf("Too many layers of recursive module executions, max allowed(%d), please fix your recursive call", maxLayers))
								}
							}()

							mTasker.ExecTask(taskname, callerVars, true)
							TaskerStack.Pop()
							os.Chdir(cwd)
						} else {
							//TODO: put the reasoning into the doco: not to auto update to avoid evil code injection problem
							u.InvalidAndExit(u.Spf("module dir: [%s] does not exist under: [%s]\n", mod.Dir, cwd), "double check if you have change your module configuration, then you will probably need to update module again")
						}
					}()
				} else {
					u.LogWarn("locating module name failed", u.Spf("module name: [%s] does not exist", modname))
					TaskerRuntime().Tasker.ListMainModules()
				}

			} else {
				callerName := TaskerRuntime().Tasker.Config.ModuleName
				u.InvalidAndExit(u.Spf("caller Module [%s] is not configured,", callerName), u.Spf("module: [%s], task: [%s]", modname, taskname))
			}
		}

	}

}

func (t *Tasker) ExecTask(taskname string, callerVars *core.Cache, isExternalCall bool) {
	found := false
	for idx, task := range *t.Tasks {
		if taskname == task.Name {
			u.Pfvvvv("  located task-> %d [%s]: \n", idx+1, task.Name)

			var ctxCallerTaskname string

			//u.Ptmpdebug("RRR", TaskerStack.GetLen())

			if isExternalCall {
				ctxCallerTaskname = "TODO: Main Caller Taskname"
			} else {
				if IsAtRootTaskLevel() {
					ctxCallerTaskname = taskname
				} else {
					ctxCallerTaskname = TaskRuntime().TasknameLayered
				}
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
					Taskname:           taskname,
					TaskVars:           core.NewCache(),
					IsCalledExternally: isExternalCall,
				}

				u.Pdebugvvvvvvv(callerVars)
				if isExternalCall {
					var passinvars core.Cache
					passinvars = deepcopy.Copy(*t.RuntimeVarsAndDvarsMerged).(core.Cache)
					mergo.Merge(&passinvars, callerVars, mergo.WithOverride)
					rtContext.ExecbaseVars = &passinvars
					rtContext.TasknameLayered = u.Spf("%s/%s", "TODO: Main Caller Taskname", taskname)
				} else {
					if IsAtRootTaskLevel() {
						rtContext.ExecbaseVars = t.RuntimeVarsAndDvarsMerged
						rtContext.TasknameLayered = taskname
					} else {
						rtContext.ExecbaseVars = callerVars
						rtContext.TasknameLayered = u.Spf("%s/%s", TaskRuntime().TasknameLayered, taskname)
					}
				}

				u.Pdebugvvvvvvv(rtContext.ExecbaseVars)

				func() {
					UpRunTimeVars.Put(UP_RUNTIME_TASK_LAYER_NUMBER, TaskerRuntime().Tasker.TaskStack.GetLen())
					TaskerRuntime().Tasker.TaskStack.Push(&rtContext)
					u.Pvvvv("Executing task stack layer:", TaskerRuntime().Tasker.TaskStack.GetLen())
					maxLayers, err := strconv.Atoi(ConfigRuntime().MaxCallLayers)
					u.LogErrorAndExit("evaluate max task stack layer", err, "please setup max MaxCallLayers correctly")

					if maxLayers != 0 && TaskerRuntime().Tasker.TaskStack.GetLen() > maxLayers {
						u.InvalidAndExit("Task exec stack layer check:", u.Spf("Too many layers of task executions, max allowed(%d), please fix your recursive call", maxLayers))
					}
				}()

				steps.Exec(false)

				returnVars := TaskRuntime().ReturnVars

				TaskerRuntime().Tasker.TaskStack.Pop()

				func() {
					//this will ensure the local caller vars are synced with return values, typically useful for chained tasks in call func
					if returnVars != nil {
						mergo.Merge(callerVars, returnVars, mergo.WithOverride)
					}

					if isExternalCall {
						if returnVars != nil {
							callerExecBaseVars := TaskerRuntime().TaskerCaller.TaskStack.GetTop().(*TaskRuntimeContext).ExecbaseVars
							mergo.Merge(callerExecBaseVars, returnVars, mergo.WithOverride)
						}
					} else {
						if !IsAtRootTaskLevel() && returnVars != nil {
							mergo.Merge(TaskRuntime().ExecbaseVars, returnVars, mergo.WithOverride)
						} else if IsAtRootTaskLevel() && returnVars != nil {
							mergo.Merge(t.RuntimeVarsAndDvarsMerged, returnVars, mergo.WithOverride)
						}
					}

				}()

			}()

		}
	}

	if !found {
		u.Pferror("Task %s is not defined!", taskname)
		t.ListTasks()
		u.InvalidAndExit("Task call failed", "Task does not exist")
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
		refdir := ConfigRuntime().RefDir

		if task.Ref != "" {
			if task.RefDir != "" {
				rawdir := task.RefDir
				refdir = Render(rawdir, t.RuntimeVarsAndDvarsMerged)
			}

			rawref := task.Ref
			ref := Render(rawref, t.RuntimeVarsAndDvarsMerged)

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
			tasksYmlRoot := u.YamlLoader(tasksYamlName, ConfigRuntime().RefDir, tasksYamlName)

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
	t.ScopeProfiles = &scopes

	u.LogErrorAndExit("load full scopes", err, "please assess your scope configuration carefully")
}

func (t *Tasker) loadExecProfiles() {
	eprofileData := t.TaskYmlRoot.Get("eprofiles")
	var eprofiles ExecProfiles
	err := ms.Decode(eprofileData, &eprofiles)
	t.ExecProfiles = &eprofiles

	u.LogErrorAndExit("load exec profiles", err, "please assess your exec profiles configuration carefully")
}

func (t *Tasker) getExecProfile(pname string) *ExecProfile {
	var ep *ExecProfile
	if t.ExecProfiles != nil {
		for _, p := range *t.ExecProfiles {
			if p.Name == pname {
				ep = &p
				break
			}
		}
	}
	return ep
}

func (t *Tasker) loadRuntimeGlobalVars() {
	varsData := t.TaskYmlRoot.Get("vars")
	var vars core.Cache
	err := ms.Decode(varsData, &vars)
	u.LogError("loadRuntimeGlobalVars", err)
	t.RuntimeGlobalVars = &vars
}

func (t *Tasker) loadRuntimeGlobalDvars() {
	dvarsData := t.TaskYmlRoot.Get("dvars")
	var dvars Dvars
	err := ms.Decode(dvarsData, &dvars)
	u.LogErrorAndExit("loadRuntimeGlobalDvars",
		err,
		"You must fix the data type to be\n string for a dvar value and try again. possible problems:\nthe name can not be single character 'y' or 'n' ",
	)
	//dvars.ValidateAndLoading()
	t.RuntimeGlobalDvars = &dvars
}

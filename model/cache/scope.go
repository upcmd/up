// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package cache

import (
	"github.com/imdario/mergo"
	ms "github.com/mitchellh/mapstructure"
	"github.com/mohae/deepcopy"
	"github.com/spf13/viper"
	u "github.com/stephencheng/up/utils"
	"gopkg.in/yaml.v2"
	"os"
)

var (
	//expanded context only contains group and global scope, but not each instance vars
	expandedContext ExpandedContext = ExpandedContext{}

	GroupMembersList []string          = []string{}
	MemberGroupMap   map[string]string = map[string]string{}
	ScopeProfiles    *Scopes

	//this is the merged vars from within scope: global, groups level (if there is), instance varss, then global runtime vars
	RuntimeVarsMerged *Cache

	//this is the merged vars and dvars to a vars cache from within scope: global, groups level (if there is), instance varss, then global runtime vars
	//this vars should be used instead of RuntimeVarsMerged as it include both runtime vars and dvars except the local vars and dvars
	RuntimeVarsAndDvarsMerged *Cache

	RuntimeGlobalVars  *Cache
	RuntimeGlobalDvars *Dvars
)

type Scope struct {
	Name    string
	Ref     string
	Members []string
	Vars    Cache
	Dvars   Dvars
}

type Scopes []Scope

type ExpandedContext map[string]*Cache
type ContextInstances []ExpandedContext

func SetScopeProfiles(sp *Scopes) {
	ScopeProfiles = sp
}

func SetRuntimeGlobalVars(vars *Cache) {
	RuntimeGlobalVars = vars
}

func SetRuntimeGlobalDvars(dvars *Dvars) {
	RuntimeGlobalDvars = dvars
}

func procDvars(dvars *Dvars, mergeTarget *Cache) {

	for _, dvar := range *dvars {
		//convert the yaml to object
		if dvar.Flags != nil && len(dvar.Flags) != 0 {
			if u.Contains(dvar.Flags, "to_object") {
				rawyml := dvar.Rendered
				obj := new(interface{})
				err := yaml.Unmarshal([]byte(rawyml), obj)
				u.LogErrorAndExit("dvar conversion to object:", err, "please validate the ymal content")

				if dvar.Expand > 1 {
					u.InvalidAndExit("dvar validation", "multiple expand > 1 is not allowed when to_object is set")
				}
				(*mergeTarget).Put(u.Spf("%s_%s", dvar.Name, "object"), *obj)
			}
		}

	}

}

func SetRuntimeGlobalMergedWithDvars() (vars *Cache) {
	var mergedVars Cache
	mergedVars = deepcopy.Copy(*RuntimeVarsMerged).(Cache)

	//u.Ptmpdebug("xxx", RuntimeGlobalDvars)
	expandedVars := RuntimeGlobalDvars.Expand("runtime global", RuntimeVarsMerged)

	if RuntimeGlobalDvars != nil {
		mergo.Merge(&mergedVars, *expandedVars, mergo.WithOverride)
	}

	RuntimeVarsAndDvarsMerged = &mergedVars
	u.Ppmsgvvvvhint("-------runtime global final merged with dvars-------", mergedVars)

	procDvars(RuntimeGlobalDvars, RuntimeVarsAndDvarsMerged)

	return RuntimeVarsAndDvarsMerged
}

func GlobalVarsMergedWithDvars(scope *Scope) (vars *Cache) {

	var mergedVars Cache
	mergedVars = deepcopy.Copy(scope.Vars).(Cache)

	expandedVars := scope.Dvars.Expand(scope.Name, &scope.Vars)

	if scope.Dvars != nil {
		mergo.Merge(&mergedVars, *expandedVars, mergo.WithOverride)
	}

	u.Pfvvvv("scope[%s] merged: %s", scope.Name, u.Sppmsg(mergedVars))

	procDvars(&scope.Dvars, &mergedVars)

	return &mergedVars
}

func GroupVarsMergedWithDvars(scope *Scope, contextMergedVars *Cache) (vars *Cache) {

	var mergedVars Cache
	mergedVars = deepcopy.Copy(scope.Vars).(Cache)

	expandedVars := scope.Dvars.Expand(scope.Name, contextMergedVars)

	if scope.Dvars != nil {
		mergo.Merge(&mergedVars, *expandedVars, mergo.WithOverride)
	}

	u.Pfvvvv("scope[%s] merged: %s", scope.Name, u.Sppmsg(mergedVars))

	procDvars(&scope.Dvars, &mergedVars)

	return &mergedVars
}

func loadRefVars(yamlroot *viper.Viper) *Cache {
	scopesData := yamlroot.Get("vars")
	vars := Cache{}
	err := ms.Decode(scopesData, &vars)
	u.Dvvvvv(vars)
	u.LogError("load ref vars", err)
	return &vars
}

/*
Get the merged vars for specific scope instance
Validate the scopes
1. for the scope name equal to global, there should be no value for members, otherwise errors
2. for the scope with group members, it is a group itself
3. for the scope with no members and name is not global, it is a final instance
*/

func (ss *Scopes) InitContextInstances() {

	//validation
	for idx, s := range *ss {

		if s.Ref != "" && s.Vars != nil {
			u.LogError("verify scope ref and member coexistence", "ref and members can not both exist")
			u.Dvvvvv(s)
			os.Exit(-1)
		}
		if s.Ref != "" {
			yamlvarsroot := u.YamlLoader("ref vars", u.CoreConfig.TaskDir, s.Ref)
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
				u.LogError("scope expand", "global scope should not contains members")
				os.Exit(-1)
			}
			globalScope = &(*ss)[idx]
		}
	}

	//expand dvars into global scope's vars space
	globalvarsMergedWithDvars := GlobalVarsMergedWithDvars(globalScope)
	for idx, s := range *ss {
		if s.Members != nil {
			for _, m := range s.Members {
				if u.Contains(GroupMembersList, m) {
					u.LogError("scope expand", u.Spfv("duplicated member: %s\n", m))
					os.Exit(-1)
				}
				GroupMembersList = append(GroupMembersList, m)
				MemberGroupMap[m] = s.Name
			}

			var groupvars Cache = deepcopy.Copy(*globalvarsMergedWithDvars).(Cache)
			mergo.Merge(&groupvars, s.Vars, mergo.WithOverride)

			//expand dvars into group scope's vars space
			groupScope := &(*ss)[idx]
			var groupvarsMergedWithDvars *Cache = GroupVarsMergedWithDvars(groupScope, &groupvars)

			expandedContext[s.Name] = groupvarsMergedWithDvars
			//u.Ptmpdebug("group merged vars", s.Name, *groupvarsMergedWithDvars)
		}
	}

	//u.Ppmsgvvvvhint("999", expandedContext)

	expandedContext["global"] = globalvarsMergedWithDvars
	ListContextInstances()
}

func ListContextInstances() {
	u.Pvvvv("---------group vars----------")
	for k, v := range expandedContext {
		u.Pfvvvv("%s: %s", k, u.Sppmsg(*v))
	}
	u.Pfvvvv("groups members:%s\n", GroupMembersList)

}

//get instance vars from scope definition, eg dev
func (ss *Scopes) GetInstanceVars(instanceName string) *Cache {
	for _, s := range *ss {
		if s.Name == instanceName {
			return &s.Vars
		}
	}

	return nil
}

/*
This will generate a one off vars merged from top level down to runtime
global and merge them all together,the result vars will be used to finally
merge with local func vars to be used in runtime execution time

pass in runtime id, if runtime id is in member list, eg dev -> nonprod
then merge runtimevars to group(nonprod)'s varss,

if runtime id (nonname) is not in member list,
then merge runtimevars to global varss,

*/
func SetRuntimeVarsMerged(runtimeid string) *Cache {
	var runtimevars Cache
	runtimevars = deepcopy.Copy(*expandedContext["global"]).(Cache)

	if u.Contains(GroupMembersList, runtimeid) {
		groupname := MemberGroupMap[runtimeid]
		mergo.Merge(&runtimevars, *expandedContext[groupname], mergo.WithOverride)

		instanceVars := ScopeProfiles.GetInstanceVars(runtimeid)
		if instanceVars != nil {
			mergo.Merge(&runtimevars, instanceVars, mergo.WithOverride)
		}

	}

	mergo.Merge(&runtimevars, *RuntimeGlobalVars, mergo.WithOverride)

	u.Pfvvvv("merged[ %s ] runtime vars:", runtimeid)
	u.Ppmsgvvvv(runtimevars)
	u.Dvvvvv(runtimevars)

	RuntimeVarsMerged = &runtimevars
	return &runtimevars
}

/*
merge localvars to above RuntimeVarsMerged to get final runtime exec vars
the localvars is the vars in the step
*/
func GetRuntimeExecVars(mark string, localvars *Cache) *Cache {
	var execvars Cache
	execvars = deepcopy.Copy(*RuntimeVarsAndDvarsMerged).(Cache)

	if localvars != nil {
		mergo.Merge(&execvars, *localvars, mergo.WithOverride)

		u.Pfvvvv("current exec runtime[%s] vars:", mark)
		u.Ppmsgvvvv(execvars)
		u.Dvvvvv(execvars)
	}
	return &execvars
}


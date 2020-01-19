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

	"os"
)

var (
	//expanded context only contains group and global scope, but not each instance vars
	expandedContext   ExpandedContext   = ExpandedContext{}
	GroupMembersList  []string          = []string{}
	MemberGroupMap    map[string]string = map[string]string{}
	ScopeProfiles     *Scopes
	RuntimeVarsMerged *Cache
	RuntimeGlobalVars *Cache
)

type Scope struct {
	Name    string
	Ref     string
	Members []string
	Vars    Cache
}

type Scopes []Scope

type ExpandedContext map[string]Cache
type ContextInstances []ExpandedContext

func SetScopeProfiles(sp *Scopes) {
	ScopeProfiles = sp
}

func SetRuntimeGlobalVars(vars *Cache) {
	RuntimeGlobalVars = vars
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
	var globalvars Cache

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

	for _, s := range *ss {
		if s.Name == "global" {
			if s.Members != nil {
				u.LogError("scope expand", "global scope should not contains members")
				os.Exit(-1)
			}
			globalvars = s.Vars
		}
	}

	for _, s := range *ss {
		if s.Members != nil {
			for _, m := range s.Members {
				if u.Contains(GroupMembersList, m) {
					u.LogError("scope expand", u.Spfv("duplicated member: %s\n", m))
					os.Exit(-1)
				}
				GroupMembersList = append(GroupMembersList, m)
				MemberGroupMap[m] = s.Name
			}

			var groupvars Cache = deepcopy.Copy(globalvars).(Cache)
			mergo.Merge(&groupvars, s.Vars, mergo.WithOverride)
			expandedContext[s.Name] = groupvars
		}
	}

	expandedContext["global"] = globalvars
	ListContextInstances()
}

func ListContextInstances() {
	u.Pvvvv("---------group vars----------")
	for k, v := range expandedContext {
		u.Pfvvvv("%s: %s", k, u.Spp(v))
		//u.Dvvvvv(k, v)
	}
	u.Pfvvvv("groups members:%s\n", GroupMembersList)

}

//get instance vars, eg dev
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
	if u.Contains(GroupMembersList, runtimeid) {
		groupname := MemberGroupMap[runtimeid]
		runtimevars = deepcopy.Copy(expandedContext[groupname]).(Cache)

		instanceVars := ScopeProfiles.GetInstanceVars(runtimeid)
		if instanceVars != nil {
			mergo.Merge(&runtimevars, instanceVars, mergo.WithOverride)
		}
	} else {
		runtimevars = deepcopy.Copy(expandedContext["global"]).(Cache)
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
*/
func GetRuntimeExecVars(mark string, localvars *Cache) *Cache {
	var execvars Cache
	execvars = deepcopy.Copy(*RuntimeVarsMerged).(Cache)

	if localvars != nil {
		mergo.Merge(&execvars, *localvars, mergo.WithOverride)

		u.Pfvvvv("current exec runtime[%s] vars:", mark)
		u.Ppmsgvvvv(execvars)
		u.Dvvvvv(execvars)
	}
	return &execvars
}


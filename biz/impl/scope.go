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
	"github.com/spf13/viper"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
	"io/ioutil"
	"os"
)

var (
	//expanded context only contains group and global scope, but not each instance vars
	expandedContext ExpandedContext = ExpandedContext{}

	GroupMembersList []string          = []string{}
	MemberGroupMap   map[string]string = map[string]string{}
	ScopeProfiles    *Scopes

	//this is the merged vars from within scope: global, groups level (if there is), instance varss, then global runtime vars
	RuntimeVarsMerged *core.Cache

	//this is the merged vars and dvars to a vars cache from within scope: global, groups level (if there is), instance varss, then global runtime vars
	//this vars should be used instead of RuntimeVarsMerged as it include both runtime vars and dvars except the local vars and dvars
	RuntimeVarsAndDvarsMerged *core.Cache

	RuntimeGlobalVars  *core.Cache
	RuntimeGlobalDvars *Dvars
)

type Scope struct {
	Name    string
	Ref     string
	RefDir  string
	Members []string
	Vars    core.Cache
	Dvars   Dvars
}

type Scopes []Scope

type ExpandedContext map[string]*core.Cache
type ContextInstances []ExpandedContext

//clear up everything in scope and cache
func Unset() {
	expandedContext = ExpandedContext{}
	GroupMembersList = []string{}
	MemberGroupMap = map[string]string{}
	ScopeProfiles = nil
	RuntimeVarsMerged = nil
	RuntimeVarsAndDvarsMerged = nil
	RuntimeGlobalVars = nil
	RuntimeGlobalDvars = nil
}

func SetScopeProfiles(sp *Scopes) {
	ScopeProfiles = sp
}

func SetRuntimeGlobalVars(vars *core.Cache) {
	RuntimeGlobalVars = vars
}

func SetRuntimeGlobalDvars(dvars *Dvars) {
	RuntimeGlobalDvars = dvars
}

func DecryptAndRegister(securetag *u.SecureSetting, dvar *Dvar, contextVars *core.Cache, mergeTarget *core.Cache) {
	s := securetag

	if s == nil {
		u.InvalidAndExit("check secure setting", "secure setting has to be explicit in dvar secure node, or as a default setting in upconfig.yml")
	}
	var encryptionkey string
	if s.KeyRef != "" {
		data, err := ioutil.ReadFile(s.KeyRef)
		u.LogErrorAndExit("load secure key from ref file", err, "please fix file loading problem")
		encryptionkey = string(data)
	}

	if s.Key != "" {
		u.Ptmpdebug("33", contextVars, s.Key)
		encryptionkey = (*contextVars).Get(s.Key).(string)
	}

	encrypted := dvar.Rendered

	if encrypted != "" && encryptionkey != "" {
		data := map[string]string{"enc_key": encryptionkey, "encrypted": encrypted}
		decrypted := Render("{{ decryptAES .enc_key .encrypted}}", data)
		secureName := u.Spf("%s_%s", "secure", dvar.Name)
		(*mergeTarget).Put(secureName, decrypted)
	} else {
		u.InvalidAndExit("dvar decrypt", u.Spf("please double check secure settings for [%s]", dvar.Name))
	}

}

func SetRuntimeGlobalMergedWithDvars() (vars *core.Cache) {
	var mergedVars core.Cache
	mergedVars = deepcopy.Copy(*RuntimeVarsMerged).(core.Cache)

	expandedVars := RuntimeGlobalDvars.Expand("runtime global", RuntimeVarsMerged)

	if RuntimeGlobalDvars != nil {
		mergo.Merge(&mergedVars, *expandedVars, mergo.WithOverride)
	}

	RuntimeVarsAndDvarsMerged = &mergedVars
	u.Ppmsgvvvvhint("-------runtime global final merged with dvars-------", mergedVars)
	//procDvars(RuntimeGlobalDvars, RuntimeVarsAndDvarsMerged)

	return RuntimeVarsAndDvarsMerged
}

func GlobalVarsMergedWithDvars(scope *Scope) (vars *core.Cache) {
	return VarsMergedWithDvars(scope.Name, &scope.Vars, &scope.Dvars, &(scope.Vars))
}

func ScopeVarsMergedWithDvars(scope *Scope, contextMergedVars *core.Cache) *core.Cache {
	return VarsMergedWithDvars(scope.Name, &scope.Vars, &scope.Dvars, contextMergedVars)
}

/*
given vars as base vars space to expand from, expand dvars against contextVars
*/
func VarsMergedWithDvars(mark string, baseVars *core.Cache, dvars *Dvars, contextMergedVars *core.Cache) *core.Cache {
	var mergedVars core.Cache
	mergedVars = deepcopy.Copy(*baseVars).(core.Cache)

	if dvars != nil {
		expandedVars := dvars.Expand(mark, contextMergedVars)
		mergo.Merge(&mergedVars, expandedVars, mergo.WithOverride)
	}

	u.Pfvvvvv("scope[%s] merged: %s", mark, u.Sppmsg(mergedVars))

	//procDvarsNew(dvars, &mergedVars)

	return &mergedVars
}

func loadRefVars(yamlroot *viper.Viper) *core.Cache {
	scopesData := yamlroot.Get("vars")
	vars := core.Cache{}
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
				u.LogError("scope expand", "global scope should not contains members")
				os.Exit(-1)
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
				if u.Contains(GroupMembersList, m) {
					u.LogError("scope expand", u.Spfv("duplicated member: %s\n", m))
					os.Exit(-1)
				}
				GroupMembersList = append(GroupMembersList, m)
				MemberGroupMap[m] = s.Name
			}

			var groupvars core.Cache = deepcopy.Copy(*globalvarsMergedWithDvars).(core.Cache)
			mergo.Merge(&groupvars, s.Vars, mergo.WithOverride)

			//expand dvars into group scope's vars space
			groupScope := &(*ss)[idx]
			var groupvarsMergedWithDvars *core.Cache = ScopeVarsMergedWithDvars(groupScope, &groupvars)

			expandedContext[s.Name] = groupvarsMergedWithDvars
		}
	}

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
func (ss *Scopes) GetInstanceVars(instanceName string) *core.Cache {
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

This has chained dvar expansion through global to group then to instance level
and finally merge with global var, except the global dvars
*/
func SetRuntimeVarsMerged(runtimeid string) *core.Cache {
	u.Pf("module: [%s] instance id: [%s]\n", ConfigRuntime().ModuleName, runtimeid)
	var runtimevars core.Cache
	runtimevars = deepcopy.Copy(*expandedContext["global"]).(core.Cache)

	if u.Contains(GroupMembersList, runtimeid) {
		groupname := MemberGroupMap[runtimeid]
		mergo.Merge(&runtimevars, *expandedContext[groupname], mergo.WithOverride)

		instanceVars := ScopeProfiles.GetInstanceVars(runtimeid)
		if instanceVars != nil {
			mergo.Merge(&runtimevars, instanceVars, mergo.WithOverride)
		}

	}

	//merge dvars for the instance
	var instanceScope *Scope
	for idx, s := range *ScopeProfiles {
		if s.Name == runtimeid {
			instanceScope = &(*ScopeProfiles)[idx]
		}
	}

	var instanceVarsMergedWithDvars *core.Cache
	if instanceScope != nil {
		instanceVarsMergedWithDvars = VarsMergedWithDvars(instanceScope.Name, &instanceScope.Vars, &instanceScope.Dvars, &runtimevars)
		//merge back the expanded merged scope vars and dvars
		mergo.Merge(&runtimevars, *instanceVarsMergedWithDvars, mergo.WithOverride)
	}

	//merge with global vars
	mergo.Merge(&runtimevars, *RuntimeGlobalVars, mergo.WithOverride)

	u.Pfvvvv("merged[ %s ] runtime vars:", runtimeid)
	u.Ppmsgvvvv(runtimevars)
	u.Dvvvvv(runtimevars)

	RuntimeVarsMerged = &runtimevars
	return &runtimevars
}


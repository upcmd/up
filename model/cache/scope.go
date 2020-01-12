// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package cache

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/imdario/mergo"
	"github.com/mohae/deepcopy"
	u "github.com/stephencheng/up/utils"
	"os"
)

var (
	contextInstances *ContextInstances
	GroupMembersList []string          = []string{}
	MemberGroupMap   map[string]string = map[string]string{}
)

type Scope struct {
	Name    string
	Members []string
	Vars    Cache
}

type Scopes []Scope

type ContextInstance map[string]Cache
type ContextInstances []ContextInstance

/*
Get the merged vars for specific scope instance
Validate the scopes
1. for the scope name equal to global, there should be no value for members, otherwise errors
2. for the scope with group members, it is a group itself
3. for the scope with no members and name is not global, it is a final instance
*/

func (ss Scopes) InitContextInstances() {
	var groupContextInstances = ContextInstances{}
	var globalvars Cache

	for _, s := range ss {
		if s.Name == "global" {
			if s.Members != nil {
				u.LogError("scope expand", "global scope should not contains members")
				os.Exit(-1)
			}
			globalvars = s.Vars
		}
	}

	u.Pfv("global address: %p\n", &globalvars)

	for _, s := range ss {
		if s.Members != nil {
			for _, m := range s.Members {
				if u.Contains(GroupMembersList, m) {
					u.LogError("scope expand", u.Spfv("duplicated member: %s\n", m))
					os.Exit(-1)
				}
				GroupMembersList = append(GroupMembersList, m)
				MemberGroupMap[m] = s.Name
			}

			u.P("-------group name:", s.Name)
			var groupvars Cache = deepcopy.Copy(globalvars).(Cache)
			u.Pfv("groupvars address: %p \n", &groupvars)
			spew.Dump("group vars base:", groupvars)
			spew.Dump("group vars:", s.Vars)
			mergo.Merge(&groupvars, s.Vars, mergo.WithOverride)
			spew.Dump("merged group vars:", groupvars)
			groupContextInstance := ContextInstance{s.Name: groupvars}
			groupContextInstances = append(groupContextInstances, groupContextInstance)
		}
	}

	groupContextInstances = append(groupContextInstances, ContextInstance{"global": globalvars})

	contextInstances = &groupContextInstances
	ListContextInstances()
}

func ListContextInstances() {
	u.P("---------group vars----------")
	for _, gci := range *contextInstances {
		spew.Dump(gci)
		u.P("")
	}
	u.P(GroupMembersList)

}

/*pass in runtime id, if runtime id is in member list, eg dev -> nonprod
then merge runtimevars to group(nonprod)'s varss,
then merge localvars to above merged result to get final runtime vars

if runtime id (nonname) is not in member list,
then merge runtimevars to global varss,
then merge localvars to above merged result to get final runtime vars
*/
func GetRuntimeInstanceVars(runtimeid string, runtimevars Cache, localvars Cache) Cache {
	if runtimeid == "noname" {
		//var globalvars Cache = deepcopy.Copy(contextInstances["global"]).(Cache)
	}
	return nil
}


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
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
	"io/ioutil"
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

type ExecProfile struct {
	Name     string
	Ref      string
	RefDir   string
	Instance string
	Evars    EnvVars
	//optional
	Taskname string
	Pure     bool
	Verbose  string
}

type ExecProfiles []ExecProfile

func DecryptAndRegister(securetag *u.SecureSetting, dvar *Dvar, contextVars *core.Cache, mergeTarget *core.Cache) {
	s := securetag

	if s == nil {
		u.InvalidAndPanic("check secure setting", "secure setting has to be explicit in dvar secure node, or as a default setting in upconfig.yml")
	}
	var encryptionkey string
	if s.KeyRef != "" {
		data, err := ioutil.ReadFile(s.KeyRef)
		u.LogErrorAndExit("load secure key from ref file", err, "please fix file loading problem")
		encryptionkey = string(data)
	}

	if s.Key != "" {
		//use vault as first priority
		opt := GetVault().Get(s.Key)
		if opt == nil {
			opt = (*contextVars).Get(s.Key)
		}
		if opt != nil {
			encryptionkey = opt.(string)
		}
	}

	encrypted := dvar.Rendered
	if encrypted != "" && encryptionkey != "" {
		data := map[string]string{"enc_key": encryptionkey, "encrypted": encrypted}
		decrypted := Render("{{ decryptAES .enc_key .encrypted}}", data)
		secureName := u.Spf("%s_%s", "secure", dvar.Name)
		(*mergeTarget).Put(secureName, decrypted)
	} else {
		u.InvalidAndPanic("dvar decrypt", u.Spf("please double check secure settings for [%s]\nyou might need to associate an instance id or an exec profile", dvar.Name))
	}

}

func Decrypt(securetag *u.SecureSetting, dvar *Dvar, contextVars *core.Cache) string {
	s := securetag
	var decrypted string
	if s == nil {
		u.InvalidAndPanic("check secure setting", "secure setting has to be explicit in dvar secure node, or as a default setting in upconfig.yml")
	}
	var encryptionkey string
	if s.KeyRef != "" {
		data, err := ioutil.ReadFile(s.KeyRef)
		u.LogErrorAndExit("load secure key from ref file", err, "please fix file loading problem")
		encryptionkey = string(data)
	}

	if s.Key != "" {
		encryptionkey = (*contextVars).Get(s.Key).(string)
	}

	encrypted := dvar.Rendered

	if encrypted != "" && encryptionkey != "" {
		data := map[string]string{"enc_key": encryptionkey, "encrypted": encrypted}
		decrypted = Render("{{ decryptAES .enc_key .encrypted}}", data)
	} else {
		u.InvalidAndPanic("dvar decrypt", u.Spf("please double check secure settings for [%s]\nyou might need to associate an instance id or an exec profile", dvar.Name))
	}
	return decrypted
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

func loadRefEvars(yamlroot *viper.Viper) *EnvVars {
	evarsData := yamlroot.Get("evars")
	vars := EnvVars{}
	err := ms.Decode(evarsData, &vars)
	u.Dvvvvv(vars)
	u.LogError("load ref vars", err)
	return &vars
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

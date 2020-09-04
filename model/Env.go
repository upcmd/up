// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package model

import "github.com/upcmd/up/model/core"

var (
	venvs *core.Cache
)

type Env struct {
	Name  string
	Value string
}

type Venv []Env

type Venvs map[string][]Venv

func GetVenv(name string) Venv {
	v := getVenvs().Get(name)
	if v == nil {
		return nil
	} else {
		return v.(Venv)
	}

}

func PutVenv(name string, venv Venv) {
	getVenvs().Put(name, venv)
}

func DeleteVenv(name string) {
	getVenvs().Delete(name)
}

func getVenvs() *core.Cache {
	if venvs == nil {
		venvs = core.NewCache()
	}
	return venvs
}

// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"reflect"
)

type SecureSetting struct {
	Type   string
	Key    string
	KeyRef string
}

type Module struct {
	Repo    string
	Tag     string
	Version string
	Alias   string
	Dir     string
}

type UpConfig struct {
	Version       string
	RefDir        string
	TaskFile      string
	Verbose       string
	ModuleName    string
	MaxCallLayers string
	Secure        *SecureSetting
	Modules       *[]Module
}

func (cfg *UpConfig) SetVerbose(cmdV string) {
	if cmdV != "" {
		cfg.Verbose = cmdV
	}
}

func (cfg *UpConfig) SetRefdir(refdir string) {
	if refdir != "" {
		cfg.RefDir = refdir
	}
}

func (cfg *UpConfig) SetTaskfile(taskfile string) {
	if taskfile != "" {
		cfg.TaskFile = taskfile
	}
}

func (cfg *UpConfig) ShowCoreConfig(mark string) {
	e := reflect.ValueOf(cfg).Elem()
	et := reflect.Indirect(e).Type()
	fmt.Printf("%s config:\n", mark)
	for i := 0; i < e.NumField(); i++ {
		if f := e.Field(i); f.Kind() == reflect.String {
			fname := et.Field(i).Name
			fmt.Printf("%20s -> %s\n", fname, f.String())
		}
	}
}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"os"
	"path"
	"reflect"
)

type SecureSetting struct {
	Type   string
	Key    string
	KeyRef string
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

func (cfg *UpConfig) GetWorkdirOld() (wkdir string) {
	cwd, err := os.Getwd()
	if err != nil {
		LogErrorAndPanic("SetAbsWorkdir", err, "working directory error")
	}

	if cfg.WorkDir == "cwd" {
		wkdir = cwd
	} else if cfg.WorkDir == "refdir" {
		//assume refdir is relative path
		abpath := path.Join(cwd, cfg.RefDir)
		if _, err := os.Stat(abpath); !os.IsNotExist(err) {
			wkdir = abpath
		} else {
			if _, err := os.Stat(cfg.RefDir); !os.IsNotExist(err) {
				wkdir = cfg.RefDir
			}
		}
	} else {
		InvalidAndPanic("SetAbsWorkdir", "Work dir setup is not proper")
	}
	return
}

//return abs path
//this must be called before tasker changing dir
func (cfg *UpConfig) SetAbsWorkdir() {
	var wkdir string
	cwd, err := os.Getwd()
	if err != nil {
		LogErrorAndPanic("SetAbsWorkdir", err, "working directory error")
	}

	if cfg.WorkDir == "cwd" {
		wkdir = cwd
	} else if cfg.WorkDir == "refdir" {
		if path.IsAbs(cfg.RefDir) {
			if _, err := os.Stat(cfg.RefDir); !os.IsNotExist(err) {
				wkdir = cfg.RefDir
			}
		} else {
			abspath := path.Clean(path.Join(cwd, cfg.RefDir))
			if _, err := os.Stat(abspath); !os.IsNotExist(err) {
				wkdir = abspath
			}
		}
	} else {
		InvalidAndPanic("SetAbsWorkdir", "Work dir setup is not proper")
	}

	cfg.AbsWorkDir = wkdir
}

func (cfg *UpConfig) SetWorkdir(workdir string) {
	if workdir != "" {
		if Contains([]string{"cwd", "refdir"}, workdir) {
			cfg.WorkDir = workdir
		}
	} else {
		cfg.WorkDir = ""
	}
}

func (cfg *UpConfig) SetTaskfile(taskfile string) {
	if taskfile != "" {
		cfg.TaskFile = taskfile
	}
}

func (cfg *UpConfig) SetModulename(modulename string) {
	if modulename != "" {
		cfg.ModuleName = modulename
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

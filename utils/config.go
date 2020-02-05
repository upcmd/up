// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stephencheng/up/model"
	"reflect"
)

var (
	Config         *viper.Viper
	CoreConfig     *model.CoreConfig
	configYamlDir  = ""
	configYamlFile = ""
)

func SetConfigYamlDir(dir string) {
	if dir != "" {
		configYamlDir = dir
	}
}

func SetConfigYamlFile(filename string) {
	if filename != "" {
		configYamlFile = filename
	}
}

func InitConfig() {
	Config = YamlLoader("Config",
		func() (s string) {
			if configYamlDir == "" {
				s = defaults["ConfigDir"]
			} else {
				s = configYamlDir
			}
			return
		}(),
		func() (s string) {
			if configYamlFile == "" {
				s = defaults["ConfigFile"]
			} else {
				s = configYamlFile
			}
			return
		}())

	CoreConfig = GetCoreConfig()
}

//for unit test only
func SetMockConfig() {
	cfg := new(model.CoreConfig)
	CoreConfig = cfg
	CoreConfig.Verbose = "vvvv"
}

func GetCoreConfig() *model.CoreConfig {

	cfg := new(model.CoreConfig)
	err := Config.Unmarshal(cfg)

	//fmt.Printf("1025: %#v\n", Config.AllSettings())

	if err != nil {
		fmt.Println("unable to decode into struct:", err.Error())
	}

	e := reflect.ValueOf(cfg).Elem()
	et := reflect.Indirect(e).Type()

	for i := 0; i < e.NumField(); i++ {
		//currently only support string type field
		if f := e.Field(i); f.Kind() == reflect.String {
			fname := et.Field(i).Name
			if val, ok := defaults[fname]; ok {
				if f.String() == "" {
					f.SetString(val)
				}
			}
		}
	}

	return cfg
}

func SetVerbose(cmdV string) {
	if cmdV != "" {
		CoreConfig.Verbose = cmdV
	}
}

func SetTaskdir(taskdir string) {
	if taskdir != "" {
		CoreConfig.TaskDir = taskdir
	}
}

func SetTaskfile(taskfile string) {
	if taskfile != "" {
		CoreConfig.TaskFile = taskfile
	}
}

func ShowCoreConfigMsg() {
	Ppmsgvvvvhint("core config", CoreConfig)
}

func ShowCoreConfig() {
	e := reflect.ValueOf(CoreConfig).Elem()
	et := reflect.Indirect(e).Type()

	for i := 0; i < e.NumField(); i++ {
		if f := e.Field(i); f.Kind() == reflect.String {
			fname := et.Field(i).Name
			Pfvvvv("%20s -> %s\n", fname, f.String())
		}
	}

}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package tests

import (
	"fmt"
	"github.com/spf13/viper"
	//"github.com/stephencheng/up/model"
	"os"
	"path"
	"runtime"
)

var (
	ymlroot              *viper.Viper
	RootDir, TestDataDir = getDirs()
)

func getDirs() (string, string) {
	_, filename, _, _ := runtime.Caller(1)
	utilsDir := path.Dir(filename)
	rootDir := path.Join(utilsDir, "..")
	return rootDir, path.Join(rootDir, "./testdata/poc")
}

func MockLoadYml(path, filename string) *viper.Viper {

	viper.AddConfigPath(".") //test in current directory as top priority then the configured path
	viper.AddConfigPath(path)
	viper.SetConfigType("yaml")
	viper.SetConfigName(filename)

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("yml file: %s/%s.yml not found...", path, filename)
		fmt.Println("errored:", err.Error())
		os.Exit(3)
	}

	ymlroot = viper.GetViper()
	return viper.GetViper()

}

//func GetNode(node string) model.Step {
//
//	var step model.Step
//	cfgEntry := ymlroot.Sub(node)
//
//	err := cfgEntry.Unmarshal(&step)
//
//	if err != nil {
//		fmt.Println("unable to decode into struct:", err.Error())
//	}
//	return step
//}



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
	"os"
	"reflect"
)

var (
	Config     = initConfig()
	CoreConfig = GetCoreConfig()
)

func initConfig() *viper.Viper {

	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("../../testdata/poc")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("Config file:config.yml not found...", err)
		os.Exit(3)
	}
	return viper.GetViper()

}

//env config is via the scope, so not to consider this for now
//func GetCoreConfig(appEnvName string) model.CoreConfig {
//
//	var cfg model.CoreConfig
//	cfgEntry := Config.Sub(appEnvName)
//
//	err := cfgEntry.Unmarshal(&cfg)
//
//	if err != nil {
//		fmt.Println("unable to decode into struct:", err.Error())
//	}
//	return cfg
//}

func GetCoreConfig() *model.CoreConfig {

	cfg := new(model.CoreConfig)
	err := Config.Unmarshal(cfg)

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


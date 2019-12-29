// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func YamlLoader(id, path, filename string) *viper.Viper {
	Pf("loading [%s] yaml file:  %s/%s\n", id, path, filename)
	newV := viper.New()
	newV.AddConfigPath(path)
	newV.SetConfigType("yaml")
	newV.SetConfigName(filename)

	//fmt.Println(path, filename)
	err := newV.ReadInConfig()

	if err != nil {
		fmt.Printf("yml file: %s/%s.yml not found...", path, filename)
		fmt.Println("errored:", err.Error())
		LogError("Yaml loading error", err)
		os.Exit(3)
	}

	return newV
}


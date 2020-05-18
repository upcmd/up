// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package core

import (
	u "github.com/stephencheng/up/utils"
	yq "github.com/stephencheng/yq/v3/cmd"
	"gopkg.in/yaml.v2"
	"strings"
)

func ObjToYaml(obj interface{}) string {
	ymlbytes, err := yaml.Marshal(&obj)
	u.LogErrorAndExit("obj to yaml converstion", err, "yml convesion failed")
	return string(ymlbytes)
}

func YamlToObj(srcyml string) interface{} {
	if srcyml == "" {
		return ""
	}
	obj := new(interface{})
	err := yaml.Unmarshal([]byte(srcyml), obj)
	u.LogErrorAndContinue("yml to object:", err, u.Spf("please validate the ymal content\n---\n%s\n---\n", u.PrintContentWithLineNuber(srcyml)))
	return obj
}

/*
obj is a cache item
path format: a.b.c(name=fr*).value
prefix will be used to get the obj, rest will be used as yq path
*/
func GetSubObjectFromCache(cache *Cache, path string, collect bool, verboseLevel string) interface{} {
	yqresult := GetSubYmlFromCache(cache, path, collect, verboseLevel)
	obj := YamlToObj(yqresult)
	return obj
}

func GetSubObjectFromYml(ymlstr string, path string, collect bool, verboseLevel string) interface{} {
	yqresult, err := yq.UpReadYmlStr(ymlstr, path, verboseLevel, collect)
	u.LogErrorAndContinue("parse sub element in yml", err, u.Spf("please ensure correct yml query path: %s", path))
	obj := YamlToObj(yqresult)
	return obj
}

func GetSubObjectFromFile(ymlfile string, path string, collect bool, verboseLevel string) interface{} {
	yqresult, err := yq.UpReadYmlFile(ymlfile, path, verboseLevel, collect)
	u.LogErrorAndContinue("parse sub element in yml", err, u.Spf("please ensure correct yml query path: %s", path))
	obj := YamlToObj(yqresult)
	return obj
}

func GetSubYmlFromYml(ymlstr string, path string, collect bool, verboseLevel string) string {
	yqresult, err := yq.UpReadYmlStr(ymlstr, path, verboseLevel, collect)
	u.LogErrorAndContinue("parse sub element in yml", err, u.Spf("please ensure correct yml query path: %s", path))
	return yqresult
}

func GetSubYmlFromFile(ymlfile string, path string, collect bool, verboseLevel string) string {
	yqresult, err := yq.UpReadYmlFile(ymlfile, path, verboseLevel, collect)
	u.LogErrorAndContinue("parse sub element in yml", err, u.Spf("please ensure correct yml query path: %s", path))
	return yqresult
}

func GetSubYmlFromCache(cache *Cache, path string, collect bool, verboseLevel string) string {
	//obj -> yml -> yq to get node in yml -> obj
	elist := strings.Split(path, ".")
	func() {
		if elist[0] == "" {
			u.InvalidAndExit("yml path validation", u.Spf("path format is not correct, use format like:\n %s", u.Yq_read_hint))
		}
	}()
	yqpath := strings.Join(elist[1:], ".")

	cacheKey := elist[0]
	obj := cache.Get(cacheKey)
	ymlstr := ObjToYaml(obj)
	u.Dvvvvv("sub yml str")
	u.Dvvvvv(ymlstr)
	yqresult, err := yq.UpReadYmlStr(ymlstr, yqpath, verboseLevel, collect)
	u.LogErrorAndContinue("parse sub element in yml", err, u.Spf("please ensure correct yml query path: %s", yqpath))
	return yqresult
}


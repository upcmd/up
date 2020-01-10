// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package model

import (
	u "github.com/stephencheng/up/utils"
	"sync"
)

var (
	CacheMap map[string]interface{} = make(map[string]interface{})
	Mutex                           = &sync.Mutex{}
)

func Put(key string, obj interface{}) {
	CacheMap[key] = obj
}

func List() {
	for k, v := range CacheMap {
		u.Pfvvvv("		%s: %+v \n", k, v)
	}
}

func Get(key string) interface{} {
	return CacheMap[key]
}

func Update(key string, obj interface{}) {
	Mutex.Lock()
	defer Mutex.Unlock()
	CacheMap[key] = obj
	CacheMap[key+"_new?"] = true

}

func Delete(key string) {
	Mutex.Lock()
	defer Mutex.Unlock()

	if _, exists := CacheMap[key]; exists {
		delete(CacheMap, key)
	}

}

//get a cached item, return a bool flag to indicate if it is an latest updated
func SafeGet(key string) (interface{}, bool) {
	Mutex.Lock()
	defer Mutex.Unlock()
	isNew := Get(key + "_new?").(bool)
	if isNew {
		return Get(key), true
	} else {
		return Get(key), false
	}
}

//mark the obj is a obsolete item due to failed get call to get latest value
func Obsolete(key string) {
	Mutex.Lock()
	defer Mutex.Unlock()
	if _, ok := CacheMap[key]; ok {
		Put(key+"_new?", false)
	} else {
		Put(key+"_new?", true)
	}
}


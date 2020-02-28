// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package core

import (
	u "github.com/stephencheng/up/utils"
	"strings"
	"sync"
)

type Cache map[string]interface{}

var (
	cacheMap *Cache
	Mutex    = &sync.Mutex{}
)

func GetCache() *Cache {
	if cacheMap == nil {
		cacheMap = NewCache()
	}
	return cacheMap
}

func NewCache() *Cache {
	cache := Cache{}
	return &cache
}

func (c *Cache) Put(key string, obj interface{}) {
	(*c)[key] = obj
}

func (c *Cache) List() {
	for k, v := range *c {
		u.Pfvvvv("		%s: %+v \n", k, v)
	}
}

func (c *Cache) Len() int {
	return len(*c)
}

func (c *Cache) GetPrefixMatched(prefix string) *Cache {
	valueMap := NewCache()
	for k, v := range *c {
		if strings.HasPrefix(k, prefix) {
			varname := strings.Trim(k, prefix)
			valueMap.Put(varname, v)
		}
	}
	return valueMap
}

func (c *Cache) Get(key string) interface{} {
	return (*c)[key]
}

func (c *Cache) Update(key string, obj interface{}) {
	Mutex.Lock()
	defer Mutex.Unlock()
	(*c)[key] = obj
	(*c)[key+"_new?"] = true
}

func (c *Cache) Delete(key string) {
	Mutex.Lock()
	defer Mutex.Unlock()

	if _, exists := (*c)[key]; exists {
		isNewKey := key + "_new?"
		v := c.Get(isNewKey)
		if v != nil {
			delete((*c), isNewKey)
		}

		delete((*c), key)
	}
}

//get a cached item, return a bool flag to indicate if it is an latest updated
func (c *Cache) SafeGet(key string) (interface{}, bool) {
	Mutex.Lock()
	defer Mutex.Unlock()
	isNewEle := c.Get(key + "_new?")
	isNew := func() (isnew bool) {
		if isNewEle == nil {
			isnew = true
		} else {
			isnew = isNewEle.(bool)
		}
		return
	}()

	if isNew {
		return c.Get(key), true
	} else {
		return c.Get(key), false
	}
}

//mark the obj is a obsolete item due to failed get call to get latest value
func (c *Cache) Obsolete(key string) {
	Mutex.Lock()
	defer Mutex.Unlock()
	if _, ok := (*c)[key]; ok {
		c.Put(key+"_new?", false)
	} else {
		c.Put(key+"_new?", true)
	}
}


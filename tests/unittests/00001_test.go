// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package unittests

import (
	"github.com/stephencheng/up/tests"
	"path"
	"runtime"

	//"github.com/stretchr/testify/assert"
	"testing"
)

func getDirs() (string, string) {
	_, filename, _, _ := runtime.Caller(1)
	utilsDir := path.Dir(filename)
	rootDir := path.Join(utilsDir, "..")
	return rootDir, path.Join(rootDir, "./testdata/poc")
}

func Test00001(t *testing.T) {

	tests.Setup(t)
	//assert := assert.New(t)

	//svc.ExecTask("task1")

}


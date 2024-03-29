// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package model

type Task struct {
	Task    interface{} //Steps
	Desc    string
	Name    string
	Public  bool
	Ref     string
	RefDir  string
	Finally interface{}
	Rescue  bool
}

type Tasks []Task

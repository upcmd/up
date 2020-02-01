// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package template

import (
	"bytes"
	u "github.com/stephencheng/up/utils"
	"text/template"
)

func Render(tstr string, obj interface{}) string {
	tname := "step_item_exec"
	t, err := template.New(tname).Parse(tstr)
	//t, err := template.New(tname).Funcs(template.FuncMap{}).Parse(tstr)
	u.LogErrorAndExit("template rendering", err, "Please fix the template issue and try again")

	var result bytes.Buffer
	//err := t.Execute(&result, obj)
	t.Execute(&result, obj)
	//u.LogError(tname, err)

	return result.String()
}


/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"github.com/fatih/color"
	"path"
)

var (
	defaults map[string]string = map[string]string{
		"RefDir":              ".",
		"WorkDir":             "cwd",
		"TaskFile":            "up.yml",
		"Verbose":             "v",
		"MaxCallLayers":       "256",
		"MaxModuelCallLayers": "256",
		"ConfigDir":           ".",
		"ConfigFile":          "upconfig.yml",
	}
	vvvv_color_printf   = color.Magenta
	verror_color_printf = color.Red
	msg_color_printf    = color.Yellow
	himsg_color_printf  = color.HiWhite
	msg_color_sprintf   = color.YellowString
	dryrun_color_print  = color.Cyan
	UpModuleDir         = ".upmodules"

	DEFAULT_CONFIG = `
version: 1.0.0
Verbose: v
MaxCallLayers: 8
MaxModuelCallLayers: 64
RefDir: .
WorkDir: cwd
TaskFile: up.yml
ConfigDir: .
ConfigFile: upconfig.yml
`

	DEFAULT_UP_TASK_YML = `
tasks:
  -
    name: Main
    desc: main entry
    task:
      -
        func: shell
        desc: main job
        do:
          - echo "hello world"
`

	Yq_read_hint = `
path format:
1. 'a.b.c'
2. 'a.*.c'
3. 'a.**.c'
4. 'a.(child.subchild==co*).c'
5. 'a.array[0].blah'
6. 'a.array[*].blah'
`
)

func GetDefaultModuleDir() string {
	return path.Join("./", UpModuleDir)
}


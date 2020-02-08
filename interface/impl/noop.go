// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	"github.com/stephencheng/up/model/cache"
	u "github.com/stephencheng/up/utils"
)

type NoopFuncAction struct {
	Do   interface{}
	Vars *cache.Cache
	Cmds *NoopCmds
}

type NoopCmd struct {
	Name string
	Desc string
	Cmd  interface{}
}

type GeneralCmd struct {
	Name  string
	Value string
}

type NoopCmds []NoopCmd

func (f *NoopFuncAction) Adapt() {
	var cmds NoopCmds

	switch f.Do.(type) {

	case []interface{}:
		err := ms.Decode(f.Do, &cmds)
		u.LogErrorAndExit("Noop adapter", err, "please fix noop command configuration")

	default:
		u.P("Not implemented!")
	}

	f.Cmds = &cmds

}

func (noopCmd *NoopCmd) runCmd(whichtype string, f func()) {
	//u.Dvvvv("111", noopCmd.Cmd)
	switch noopCmd.Cmd.(type) {
	case string:
		if whichtype == "string" {
			f()
		}

	case map[interface{}]interface{}:
		if whichtype == "map" {
			f()
		}

	default:
		u.P("Not implemented!")
	}

}

func (f *NoopFuncAction) Exec() {

	u.P("executing noop commands")
	for idx, cmdItem := range *f.Cmds {
		u.Pfv("noop cmdItem(%2d): %s (%s)\n%s\n", idx+1, cmdItem.Name, cmdItem.Desc, color.HiBlueString("%s", cmdItem.Cmd))

		switch cmdItem.Name {
		case "print":
			cmdItem.runCmd("string", func() {
				cmdRendered := cache.Render(cmdItem.Cmd.(string), f.Vars)
				u.Pfv("%s\n", color.HiGreenString("%s", cmdRendered))
			})

		case "dereg":
			cmdItem.runCmd("string", func() {
				varname := cache.Render(cmdItem.Cmd.(string), f.Vars)
				u.Pfv("deregister var: %s\n", color.HiGreenString("%s", varname))
				cache.RuntimeVarsAndDvarsMerged.Delete(varname)
				f.Vars.Delete(varname)
			})
			u.Ppmsgvvvvvhint("after reg the var - global:", cache.RuntimeVarsAndDvarsMerged)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "reg":
			cmdItem.runCmd("map", func() {
				regCmd := cmdItem.Cmd.(map[interface{}]interface{})
				var varname, varvalue string
				var localonly bool
				for k, v := range regCmd {
					if k.(string) == "name" {
						varname = v.(string)
					}
					if k.(string) == "value" {
						varvalueRaw := v.(string)
						varvalue = cache.Render(varvalueRaw, f.Vars)
					}
					if k.(string) == "localonly" {
						localonly = v.(bool)
					}
				}

				if localonly {
					f.Vars.Put(varname, varvalue)
				} else {
					cache.RuntimeVarsAndDvarsMerged.Put(varname, varvalue)
					f.Vars.Put(varname, varvalue)
				}

			})
			u.Ppmsgvvvvvhint("after reg the var - global:", cache.RuntimeVarsAndDvarsMerged)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)
		default:
			u.Pferror("warrning: check noop cmd name:(%s),%s\n", cmdItem.Name, "cmd not implemented")
		}

	}
}


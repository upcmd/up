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
	"io/ioutil"
	"path"
)

type CmdFuncAction struct {
	Do   interface{}
	Vars *cache.Cache
	Cmds *CmdCmds
}

type CmdCmd struct {
	Name string
	Desc string
	Cmd  interface{}
}

type GeneralCmd struct {
	Name  string
	Value string
}

type CmdCmds []CmdCmd

func (f *CmdFuncAction) Adapt() {
	var cmds CmdCmds

	switch f.Do.(type) {

	case []interface{}:
		err := ms.Decode(f.Do, &cmds)
		u.LogErrorAndExit("Cmd adapter", err, "please fix cmd command configuration")

	default:
		u.LogWarn("cmd", "Not implemented or void for no action!")
	}

	f.Cmds = &cmds

}

func (cmdCmd *CmdCmd) runCmd(whichtype string, f func()) {
	//u.Dvvvv("111", cmdCmd.Cmd)
	switch cmdCmd.Cmd.(type) {
	case string:
		if whichtype == "string" {
			f()
		}

	case map[interface{}]interface{}:
		if whichtype == "map" {
			f()
		}

	default:
		u.LogWarn("cmd", "Not implemented or void for no action!")
	}

}

func (f *CmdFuncAction) Exec() {

	//u.P("executing cmd commands")
	for idx, cmdItem := range *f.Cmds {
		//u.Pfv("cmd cmdItem(%2d): %s (%s)\n%s\n", idx+1, cmdItem.Name, cmdItem.Desc, color.HiBlueString("%s", cmdItem.Cmd))
		u.Pfv("cmd cmdItem(%2d): %s (%s)\n", idx+1, cmdItem.Name, cmdItem.Desc)
		u.Pfvv("%s\n", color.MagentaString("%s", cmdItem.Cmd))

		u.LogDesc("substep", cmdItem.Desc)
		switch cmdItem.Name {
		case "print":
			cmdItem.runCmd("string", func() {
				cmdRendered := cache.Render(cmdItem.Cmd.(string), f.Vars)
				u.Pfv("%s\n", color.HiGreenString("%s", cmdRendered))
			})

		case "printobj":
			u.Dvvvv(cmdItem.Cmd)
			cmdItem.runCmd("string", func() {
				objname := cache.Render(cmdItem.Cmd.(string), f.Vars)
				obj := cache.RuntimeVarsAndDvarsMerged.Get(objname)
				u.Ppfmsg(u.Spf("object: %s", objname), obj)
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

		case "readfile":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var varname, filename, dir, raw string
				var localonly bool
				for k, v := range cmd {
					switch k.(string) {
					case "reg":
						raw = v.(string)
						varname = cache.Render(raw, f.Vars)
					case "filename":
						raw = v.(string)
						filename = cache.Render(raw, f.Vars)
					case "dir":
						raw = v.(string)
						dir = cache.Render(raw, f.Vars)
					case "localonly":
						localonly = v.(bool)
					}
				}
				filepath := path.Join(dir, filename)

				content, err := ioutil.ReadFile(filepath)
				u.LogErrorAndExit("cmd readfile", err, "please fix file path and name issues")

				if localonly {
					f.Vars.Put(varname, string(content))
				} else {
					cache.RuntimeVarsAndDvarsMerged.Put(varname, string(content))
					f.Vars.Put(varname, string(content))
				}

			})

			u.Ppmsgvvvvvhint("after reg the var - global:", cache.RuntimeVarsAndDvarsMerged)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "writefile":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var content, filename, dir, raw string
				for k, v := range cmd {
					switch k.(string) {
					case "content":
						contentRaw := v.(string)
						content = cache.Render(contentRaw, f.Vars)
					case "filename":
						raw = v.(string)
						filename = cache.Render(raw, f.Vars)
					case "dir":
						raw = v.(string)
						dir = cache.Render(raw, f.Vars)
					}
				}
				filepath := path.Join(dir, filename)
				ioutil.WriteFile(filepath, []byte(content), 0644)
			})

		case "template":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var src, dest, raw string
				for k, v := range cmd {
					switch k.(string) {
					case "src":
						raw = v.(string)
						src = cache.Render(raw, f.Vars)
					case "dest":
						raw = v.(string)
						dest = cache.Render(raw, f.Vars)
					}
				}

				tbuf, err := ioutil.ReadFile(src)
				rendered := cache.Render(string(tbuf), f.Vars)
				u.LogErrorAndExit("cmd template", err, "please fix file path and name issues")
				ioutil.WriteFile(dest, []byte(rendered), 0644)
			})

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
			u.Pferror("warrning: check cmd name:(%s),%s\n", cmdItem.Name, "cmd not implemented")
		}

	}
}


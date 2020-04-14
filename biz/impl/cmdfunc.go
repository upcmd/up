// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
	yq "github.com/stephencheng/yq/v3/cmd"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

type CmdFuncAction struct {
	Do   interface{}
	Vars *core.Cache
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
	invalidTypeHint := func(typeGot string) {
		u.LogWarn("type mismatch", u.Spf("cmd name: %s -> type wanted: %s, got :%s", cmdCmd.Name, whichtype, typeGot))
	}
	switch t := cmdCmd.Cmd.(type) {
	case string:
		if whichtype == "string" {
			f()
		} else {
			invalidTypeHint("string")
		}

	case int:
		if whichtype == "int" {
			f()
		} else {
			invalidTypeHint("int")
		}

	case map[interface{}]interface{}:
		if whichtype == "map" {
			f()
		} else {
			invalidTypeHint("map")
		}

	case []interface{}:
		if whichtype == "array" {
			f()
		} else {
			invalidTypeHint("array")
		}

	default:
		u.LogWarn("cmd", u.Spf("Not implemented type(%T) or void for no action!", t))
	}

}

func (f *CmdFuncAction) Exec() {

	for idx, cmdItem := range *f.Cmds {
		if cmdItem.Cmd != nil {
			u.Pfvvvvv("%s\n", color.MagentaString("%s", cmdItem.Cmd))
		}

		taskLayerCnt := core.TaskStack.GetLen()
		u.LogDesc("substep", idx+1, taskLayerCnt, cmdItem.Name, cmdItem.Desc)
		switch cmdItem.Name {
		case "print":
			cmdItem.runCmd("string", func() {
				cmdRendered := core.Render(cmdItem.Cmd.(string), f.Vars)
				u.Pfv("%s\n", color.HiGreenString("%s", cmdRendered))
			})

		case "trace":
			cmdItem.runCmd("string", func() {
				cmdRendered := core.Render(cmdItem.Cmd.(string), f.Vars)
				u.Ptrace("Trace:", cmdRendered)
			})

		case "printobj":
			u.Dvvvv(cmdItem.Cmd)
			cmdItem.runCmd("string", func() {
				objname := core.Render(cmdItem.Cmd.(string), f.Vars)
				obj := f.Vars.Get(objname)
				u.Ppfmsg(u.Spf("object:\n %s", objname), obj)
			})

		case "dereg":
			cmdItem.runCmd("string", func() {
				varname := core.Render(cmdItem.Cmd.(string), f.Vars)
				u.Pfv("deregister var: %s\n", color.HiGreenString("%s", varname))
				core.TaskRuntime().ExecbaseVars.Delete(varname)
				f.Vars.Delete(varname)
			})
			u.Ppmsgvvvvvhint("after reg the var - contextual global:", core.TaskRuntime().ExecbaseVars)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "sleep":
			cmdItem.runCmd("int", func() {
				mscnt := cmdItem.Cmd.(int)
				u.Sleep(mscnt)
			})

		case "pause":
			pause(f.Vars)

		case "exit":
			u.GraceExit("exit", "client choose to exit")

		case "readfile":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var varname, filename, dir, raw string
				var localonly bool
				for k, v := range cmd {
					switch k.(string) {
					case "reg":
						raw = v.(string)
						varname = core.Render(raw, f.Vars)
					case "filename":
						raw = v.(string)
						filename = core.Render(raw, f.Vars)
					case "dir":
						raw = v.(string)
						dir = core.Render(raw, f.Vars)
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
					core.TaskRuntime().ExecbaseVars.Put(varname, string(content))
					f.Vars.Put(varname, string(content))
				}

			})

			u.Ppmsgvvvvvhint("after reg the var - contextual global:", core.TaskRuntime().ExecbaseVars)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "writefile":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var content, filename, dir, raw string
				for k, v := range cmd {
					switch k.(string) {
					case "content":
						contentRaw := v.(string)
						content = core.Render(contentRaw, f.Vars)
					case "filename":
						raw = v.(string)
						filename = core.Render(raw, f.Vars)
					case "dir":
						raw = v.(string)
						dir = core.Render(raw, f.Vars)
					}
				}
				filepath := path.Join(dir, filename)
				ioutil.WriteFile(filepath, []byte(content), 0644)
			})

		case "template":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var src, dest, raw, datakey, datapath, rendered string
				var data interface{}
				for k, v := range cmd {
					switch k.(string) {
					case "src":
						raw = v.(string)
						src = core.Render(raw, f.Vars)
					case "datakey":
						raw = v.(string)
						datakey = core.Render(raw, f.Vars)
						data = f.Vars.Get(datakey)
					case "datapath":
						raw = v.(string)
						datapath = core.Render(raw, f.Vars)
						data = core.GetSubObjectFromCache(f.Vars, datapath, false)
						u.Ppmsgvvvvv("sub object:", data)
					case "dest":
						raw = v.(string)
						dest = core.Render(raw, f.Vars)
					}
				}

				tbuf, err := ioutil.ReadFile(src)
				if data == nil || data == "" {
					rendered = core.Render(string(tbuf), f.Vars)
				} else {
					rendered = core.Render(string(tbuf), data)
				}

				u.LogErrorAndExit("cmd template", err, "please fix file path and name issues")
				ioutil.WriteFile(dest, []byte(rendered), 0644)
			})

		case "query":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var raw, reg, ymlkey, ymlfile, yqpath string
				var collect, localonly, ymlonly bool
				refdir := u.CoreConfig.RefDir
				var data interface{}
				for k, v := range cmd {
					switch k.(string) {
					case "ymlkey":
						raw = v.(string)
						ymlkey = core.Render(raw, f.Vars)
					case "ymlfile":
						raw = v.(string)
						ymlfile = core.Render(raw, f.Vars)
					case "refdir":
						raw = v.(string)
						refdir = core.Render(raw, f.Vars)
					case "reg":
						raw = v.(string)
						reg = core.Render(raw, f.Vars)
					case "path":
						//yqpath used as:
						//1. a yqpath ref in yml content
						//2. a yqpath ref in cached object
						raw = v.(string)
						yqpath = core.Render(raw, f.Vars)
					case "localonly":
						localonly = v.(bool)
					case "ymlonly":
						ymlonly = v.(bool)
					case "collect":
						collect = v.(bool)
					}
				}

				if yqpath == "" || reg == "" {
					u.InvalidAndExit("query cmd mandatory attribute validation", "path and reg are all mandatory and required")
				}

				if ymlkey != "" {
					ymlstr := f.Vars.Get(ymlkey).(string)
					if ymlonly {
						data = core.GetSubYmlFromYml(ymlstr, yqpath, collect)
					} else {
						data = core.GetSubObjectFromYml(ymlstr, yqpath, collect)
					}
				} else if ymlfile != "" {
					filepath := path.Join(refdir, ymlfile)
					if ymlonly {
						data = core.GetSubYmlFromFile(filepath, yqpath, collect)
					} else {
						data = core.GetSubObjectFromFile(filepath, yqpath, collect)
					}
				} else if yqpath != "" {
					//means to retrieve from cache
					if ymlonly {
						data = core.GetSubYmlFromCache(f.Vars, yqpath, collect)
					} else {
						data = core.GetSubObjectFromCache(f.Vars, yqpath, collect)
					}
				}

				u.Ppmsgvvvvvhint("data object:", data)
				if localonly {
					f.Vars.Put(reg, data)
				} else {
					core.TaskRuntime().ExecbaseVars.Put(reg, data)
					f.Vars.Put(reg, data)
				}

			})

		case "yml_delete":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var raw, ymlfile, yqpath, reg string

				refdir := u.CoreConfig.RefDir
				verbose := u.CoreConfig.Verbose
				var inplace, localonly bool
				for k, v := range cmd {
					switch k.(string) {
					case "ymlfile":
						raw = v.(string)
						ymlfile = core.Render(raw, f.Vars)
					case "refdir":
						raw = v.(string)
						refdir = core.Render(raw, f.Vars)
					case "path":
						raw = v.(string)
						yqpath = core.Render(raw, f.Vars)
					case "verbose":
						verbose = v.(string)
					case "inplace":
						inplace = v.(bool)
					case "reg":
						raw = v.(string)
						reg = core.Render(raw, f.Vars)
					case "localonly":
						localonly = v.(bool)
					}
				}

				if yqpath == "" || ymlfile == "" {
					u.InvalidAndExit("mandatory attribute validation", "ymlfile and path are mandatory and required")
				}

				if inplace == true && reg != "" {
					u.InvalidAndExit("yml_delete criteria validation", "inplace and reg are mutual exclusive")
				}

				modified, err := yq.UpDeletePathFromFile(path.Join(refdir, ymlfile), yqpath, inplace, verbose)
				u.LogErrorAndContinue("delete sub element in yml", err, u.Spf("please ensure correct yml query path: %s", yqpath))
				u.Ppmsgvvvvvhint("yml modified:", modified)

				if inplace != true && reg != "" {
					if localonly {
						f.Vars.Put(reg, modified)
					} else {
						core.TaskRuntime().ExecbaseVars.Put(reg, modified)
						f.Vars.Put(reg, modified)
					}
				}
			})

		case "yml_write":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var raw, yqpath, ymlstr, reg, value, nodevalue, modified string
				var err error

				verbose := u.CoreConfig.Verbose
				var localonly bool
				for k, v := range cmd {
					switch k.(string) {
					case "ymlstr":
						raw = v.(string)
						ymlstr = core.Render(raw, f.Vars)
					case "value":
						raw = v.(string)
						value = core.Render(raw, f.Vars)
					case "nodevalue":
						raw = v.(string)
						nodevalue = core.Render(raw, f.Vars)
					case "path":
						raw = v.(string)
						yqpath = core.Render(raw, f.Vars)
					case "verbose":
						verbose = v.(string)
					case "reg":
						raw = v.(string)
						reg = core.Render(raw, f.Vars)
					case "localonly":
						localonly = v.(bool)
					}
				}

				if ymlstr == "" || yqpath == "" || reg == "" {
					u.InvalidAndExit("mandatory attribute validation", "ymlstr, path and reg are required")
				}

				if value != "" && nodevalue != "" {
					u.InvalidAndExit("value validation", "value and nodevalue are mutual exclusive")
				}

				if value != "" {
					modified, err = yq.UpWriteNodeFromStrForSimpleValue(ymlstr, yqpath, value, verbose)
				} else if nodevalue != "" {
					modified, err = yq.UpWriteNodeFromStrForComplexValueFromYmlStr(ymlstr, yqpath, nodevalue, verbose)
				}

				u.LogErrorAndContinue("write node in yml", err, u.Spf("please ensure correct yml query path: %s", yqpath))

				u.Ppmsgvvvvvhint("yml modified:", modified)

				if localonly {
					f.Vars.Put(reg, modified)
				} else {
					core.TaskRuntime().ExecbaseVars.Put(reg, modified)
					f.Vars.Put(reg, modified)
				}

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
						varvalue = core.Render(varvalueRaw, f.Vars)
					}
					if k.(string) == "localonly" {
						localonly = v.(bool)
					}
				}

				if varname == "" {
					u.InvalidAndExit("validate varname", "the reg varname must not be empty")
				}
				if localonly {
					f.Vars.Put(varname, varvalue)
				} else {
					core.TaskRuntime().ExecbaseVars.Put(varname, varvalue)
					f.Vars.Put(varname, varvalue)
				}
			})
			u.Ppmsgvvvvvhint("after reg the var - contextual global:", core.TaskRuntime().ExecbaseVars)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "return":
			cmdItem.runCmd("array", func() {
				retNames := cmdItem.Cmd.([]interface{})
				var retName string

				if core.TaskRuntime().ReturnVars == nil {
					core.TaskRuntime().ReturnVars = core.NewCache()
				}

				for _, v := range retNames {
					rawName := v.(string)
					retName = core.Render(rawName, f.Vars)
					ret := f.Vars.Get(retName)
					if ret != nil {
						core.TaskRuntime().ReturnVars.Put(retName, f.Vars.Get(retName))
					} else {
						u.LogWarn("return validation", u.Spf("The referencing var name: (%s) not exist", retName))
					}
				}

			})
			u.Ppmsgvvvvvhint("contextual return vars:", core.TaskRuntime().ReturnVars)

		case "to_object":
			//src: a var name to get the yml content from
			//reg: a registered name to cache the variable
			//localonly: if set, then the variable will not be saved to global space
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var fromkey, src, reg string
				var localonly bool
				for k, v := range cmd {
					if k.(string) == "fromkey" {
						keyRaw := v.(string)
						fromkey = core.Render(keyRaw, f.Vars)
					}
					if k.(string) == "src" {
						srcRaw := v.(string)
						src = core.Render(srcRaw, f.Vars)
					}
					if k.(string) == "reg" {
						regRaw := v.(string)
						reg = core.Render(regRaw, f.Vars)
					}
					if k.(string) == "localonly" {
						localonly = v.(bool)
					}
				}

				srcyml := func() string {
					if src != "" && fromkey != "" {
						u.InvalidAndExit("locate yml string", "you can only use either key or src, but not both")
					}
					if src != "" {
						return src
					}
					if fromkey != "" {
						t := f.Vars.Get(fromkey)
						if t != nil {
							return t.(string)
						} else {
							u.InvalidAndExit("locate yml string", "please use a valid addressable varkey to locate a yml document")
							return ""
						}
					}
					return ""
				}()
				obj := new(interface{})
				err := yaml.Unmarshal([]byte(srcyml), obj)
				u.LogErrorAndExit("cmd to_object:", err, "please validate the ymal content")

				if localonly {
					(*f.Vars).Put(reg, *obj)
				} else {
					core.TaskRuntime().ExecbaseVars.Put(src, reg)
					(*f.Vars).Put(reg, *obj)
				}

			})
			u.Ppmsgvvvvvhint("after reg the var - contextual global:", core.TaskRuntime().ExecbaseVars)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)
		default:
			u.Pferror("warrning: check cmd name:(%s),%s\n", cmdItem.Name, "cmd not implemented")
		}

	}
}


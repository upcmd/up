// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/fatih/color"
	ms "github.com/mitchellh/mapstructure"
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
	yq "github.com/upcmd/yq/v3/cmd"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"strconv"
)

type CmdFuncAction struct {
	Do   interface{}
	Vars *core.Cache
	Cmds *CmdCmds
}

type CmdCmd struct {
	Name  string
	Desc  string
	Cmd   interface{}
	Cmdx  interface{}
	Flags []string
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

	case nil:
		if cmdCmd.Cmdx != nil {
			u.LogWarn("cmd", "temporarily deactivated")
		} else {
			u.LogWarn("cmd", "lacking detailed implementation yet")
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

		taskLayerCnt := TaskerRuntime().Tasker.TaskStack.GetLen()
		u.LogDesc("substep", idx+1, taskLayerCnt, cmdItem.Name, cmdItem.Desc)

		doFlag := func(flag string, doFlagFunc func()) {
			if cmdItem.Flags != nil && u.Contains(cmdItem.Flags, flag) {
				doFlagFunc()
			}
		}

		switch cmdItem.Name {
		case "print":
			cmdItem.runCmd("string", func() {
				cmdRendered := Render(cmdItem.Cmd.(string), f.Vars)
				u.Pfv("%s\n", color.HiGreenString("%s", cmdRendered))
			})

		case "colorprint":

			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var raw, msg, fg, bg, object string

				for k, v := range cmd {
					switch k.(string) {
					case "msg":
						raw = v.(string)
						msg = Render(raw, f.Vars)
					case "fg":
						raw = v.(string)
						fg = Render(raw, f.Vars)
					case "bg":
						raw = v.(string)
						bg = Render(raw, f.Vars)
					case "object":
						raw = v.(string)
						object = Render(raw, f.Vars)
					}
				}

				var fgcolor, bgcolor color.Attribute
				if fg != "" {
					if c, ok := u.FgColorMap[fg]; ok {
						fgcolor = c
					} else {
						fgcolor = color.FgWhite
					}
				} else {
					fgcolor = color.FgWhite
				}
				if bg != "" {
					if c, ok := u.BgColorMap[bg]; ok {
						bgcolor = c
					} else {
						bgcolor = color.BgBlack
					}
				} else {
					bgcolor = color.BgBlack
				}

				c := color.New(bgcolor, fgcolor)
				u.Pln(color.FgWhite, color.BgBlue)

				if msg != "" && object != "" {
					u.LogWarn("colorprint", "msg and object can not coexist")
				} else {
					if msg != "" {
						c.Printf("%s\n", msg)
					}

					if object != "" {
						obj := f.Vars.Get(object)
						c.Printf("object %s:\n %s", object, u.Sppmsg(obj))
					}
				}
			})

		case "trace":
			cmdItem.runCmd("string", func() {
				cmdRendered := Render(cmdItem.Cmd.(string), f.Vars)
				u.Ptrace("Trace:", cmdRendered)
			})

		case "printobj":
			u.Dvvvv(cmdItem.Cmd)
			cmdItem.runCmd("string", func() {
				objname := Render(cmdItem.Cmd.(string), f.Vars)
				obj := f.Vars.Get(objname)
				u.Ppfmsg(u.Spf("object:\n %s", objname), obj)
			})

		case "dereg":
			cmdItem.runCmd("string", func() {
				varname := Render(cmdItem.Cmd.(string), f.Vars)
				u.Pfv("deregister var: %s\n", color.HiGreenString("%s", varname))
				TaskRuntime().ExecbaseVars.Delete(varname)
				f.Vars.Delete(varname)
			})
			u.Ppmsgvvvvvhint("after reg the var - contextual global:", TaskRuntime().ExecbaseVars)
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

		case "fail":
			u.Fail("fail", "fail and exit")

		case "break":
			TaskerRuntime().Tasker.TaskBreak = true

		case "assert":
			cmdItem.runCmd("array", func() {
				conditions := cmdItem.Cmd.([]interface{})
				var condition string

				var failed bool
				for idx, v := range conditions {
					raw := v.(string)
					condition = Render(raw, f.Vars)
					succeeded, err := strconv.ParseBool(condition)
					if !succeeded {
						color.Red("%2d ASSERT FAILED: [%s]", idx+1, raw)
						failed = true
						u.LogError("Reason:", err)
					} else {
						color.Green("%2d ASSERT OK:     [%s]", idx+1, raw)
					}
				}

				if failed {
					doFlag("failfast", func() {
						u.InvalidAndExit("Assert Failed", "Failfast and STOPS here!!!")
					})
				}

			})

		case "inspect":
			cmdItem.runCmd("array", func() {
				whats := cmdItem.Cmd.([]interface{})

				for idx, v := range whats {
					what := v.(string)
					u.Pf("%2d: inspect[%s]", idx+1, v)
					switch what {
					case "exec_base_vars":
						u.Ppmsg(*TaskRuntime().ExecbaseVars)
					case "exec_vars":
						u.Ppmsg(f.Vars)
					}

				}
			})

		case "readfile":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var varname, filename, dir, raw string
				var localonly bool
				for k, v := range cmd {
					switch k.(string) {
					case "reg":
						raw = v.(string)
						varname = Render(raw, f.Vars)
					case "filename":
						raw = v.(string)
						filename = Render(raw, f.Vars)
					case "dir":
						raw = v.(string)
						dir = Render(raw, f.Vars)
					}
				}

				doFlag("localonly", func() {
					localonly = true
				})

				filepath := path.Join(dir, filename)

				content, err := ioutil.ReadFile(filepath)
				u.LogErrorAndExit("cmd readfile", err, "please fix file path and name issues")

				if localonly {
					f.Vars.Put(varname, string(content))
				} else {
					TaskRuntime().ExecbaseVars.Put(varname, string(content))
					f.Vars.Put(varname, string(content))
				}

			})

			u.Ppmsgvvvvvhint("after reg the var - contextual global:", TaskRuntime().ExecbaseVars)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "writefile":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var content, filename, dir, raw string
				for k, v := range cmd {
					switch k.(string) {
					case "content":
						contentRaw := v.(string)
						content = Render(contentRaw, f.Vars)
					case "filename":
						raw = v.(string)
						filename = Render(raw, f.Vars)
					case "dir":
						raw = v.(string)
						dir = Render(raw, f.Vars)
					}
				}
				filepath := path.Join(dir, filename)
				ioutil.WriteFile(filepath, []byte(content), 0644)
			})

		case "template":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				refdir := ConfigRuntime().RefDir
				var src, dest, raw, datakey, datapath, datafile, rendered string
				var data interface{}
				dataCnt := 0
				for k, v := range cmd {
					switch k.(string) {
					case "src":
						raw = v.(string)
						src = Render(raw, f.Vars)
					case "refdir":
						raw = v.(string)
						refdir = Render(raw, f.Vars)
					case "datafile":
						raw = v.(string)
						datafile = Render(raw, f.Vars)
						dataCnt += 1
					case "datakey":
						raw = v.(string)
						datakey = Render(raw, f.Vars)
						data = f.Vars.Get(datakey)
						dataCnt += 1
					case "datapath":
						raw = v.(string)
						datapath = Render(raw, f.Vars)
						data = core.GetSubObjectFromCache(f.Vars, datapath, false, ConfigRuntime().Verbose)
						u.Ppmsgvvvvv("sub object:", data)
						dataCnt += 1
					case "dest":
						raw = v.(string)
						dest = Render(raw, f.Vars)
					}
				}

				if dataCnt > 1 {
					u.InvalidAndExit("data validation", "only one data source is alllowed")
				}

				if datafile != "" {
					data = core.LoadObjectFromFile(path.Join(refdir, datafile))
				}

				tbuf, err := ioutil.ReadFile(src)
				if data == nil || data == "" {
					rendered = Render(string(tbuf), f.Vars)
				} else {
					rendered = Render(string(tbuf), data)
				}

				u.LogErrorAndExit("read template", err, "please fix file path and name issues")
				err = ioutil.WriteFile(dest, []byte(rendered), 0644)
				u.LogErrorAndExit("write template", err, "please fix file path and name issues")

			})

		case "query":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var raw, reg, ymlkey, ymlfile, yqpath string
				var collect, localonly, ymlonly bool
				refdir := ConfigRuntime().RefDir
				var data interface{}
				for k, v := range cmd {
					switch k.(string) {
					case "ymlkey":
						raw = v.(string)
						ymlkey = Render(raw, f.Vars)
					case "ymlfile":
						raw = v.(string)
						ymlfile = Render(raw, f.Vars)
					case "refdir":
						raw = v.(string)
						refdir = Render(raw, f.Vars)
					case "reg":
						raw = v.(string)
						reg = Render(raw, f.Vars)
					case "path":
						//yqpath used as:
						//1. a yqpath ref in yml content
						//2. a yqpath ref in cached object
						raw = v.(string)
						yqpath = Render(raw, f.Vars)
					}
				}

				doFlag("localonly", func() {
					localonly = true
				})
				doFlag("ymlonly", func() {
					ymlonly = true
				})
				doFlag("collect", func() {
					collect = true
				})

				if yqpath == "" || reg == "" {
					u.InvalidAndExit("query cmd mandatory attribute validation", "path and reg are all mandatory and required")
				}

				if ymlkey != "" {
					tmpymlstr := f.Vars.Get(ymlkey)
					if tmpymlstr == nil {
						u.InvalidAndExit("data validation", "ymlkey does not exist, please fix it")
					}
					ymlstr := tmpymlstr.(string)
					if ymlonly {
						data = core.GetSubYmlFromYml(ymlstr, yqpath, collect, ConfigRuntime().Verbose)
					} else {
						data = core.GetSubObjectFromYml(ymlstr, yqpath, collect, ConfigRuntime().Verbose)
					}
				} else if ymlfile != "" {
					filepath := path.Join(refdir, ymlfile)
					if ymlonly {
						data = core.GetSubYmlFromFile(filepath, yqpath, collect, ConfigRuntime().Verbose)
					} else {
						data = core.GetSubObjectFromFile(filepath, yqpath, collect, ConfigRuntime().Verbose)
					}
				} else if yqpath != "" {
					//means to retrieve from cache
					if ymlonly {
						data = core.GetSubYmlFromCache(f.Vars, yqpath, collect, ConfigRuntime().Verbose)
					} else {
						data = core.GetSubObjectFromCache(f.Vars, yqpath, collect, ConfigRuntime().Verbose)
					}
				}

				if localonly {
					f.Vars.Put(reg, data)
				} else {
					TaskRuntime().ExecbaseVars.Put(reg, data)
					f.Vars.Put(reg, data)
				}

			})

		case "yml_delete":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var raw, ymlfile, yqpath, reg string

				refdir := ConfigRuntime().RefDir
				verbose := ConfigRuntime().Verbose
				var inplace, localonly bool
				for k, v := range cmd {
					switch k.(string) {
					case "ymlfile":
						raw = v.(string)
						ymlfile = Render(raw, f.Vars)
					case "refdir":
						raw = v.(string)
						refdir = Render(raw, f.Vars)
					case "path":
						raw = v.(string)
						yqpath = Render(raw, f.Vars)
					case "verbose":
						verbose = v.(string)
					case "reg":
						raw = v.(string)
						reg = Render(raw, f.Vars)
					}
				}

				doFlag("localonly", func() {
					localonly = true
				})
				doFlag("inplace", func() {
					inplace = true
				})

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
						TaskRuntime().ExecbaseVars.Put(reg, modified)
						f.Vars.Put(reg, modified)
					}
				}
			})

		case "yml_write":
			cmdItem.runCmd("map", func() {
				cmd := cmdItem.Cmd.(map[interface{}]interface{})
				var raw, yqpath, ymlstr, reg, value, nodevalue, modified string
				var err error

				verbose := ConfigRuntime().Verbose
				var localonly bool
				for k, v := range cmd {
					switch k.(string) {
					case "ymlstr":
						raw = v.(string)
						ymlstr = Render(raw, f.Vars)
					case "value":
						raw = v.(string)
						value = Render(raw, f.Vars)
					case "nodevalue":
						raw = v.(string)
						nodevalue = Render(raw, f.Vars)
					case "path":
						raw = v.(string)
						yqpath = Render(raw, f.Vars)
					case "verbose":
						verbose = v.(string)
					case "reg":
						raw = v.(string)
						reg = Render(raw, f.Vars)
					}
				}
				doFlag("localonly", func() {
					localonly = true
				})

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

				u.LogErrorAndContinue("write node in yml", err, u.Spf("please ensure correct yml query path: %s\nand check yml content validity:\n%s\n", yqpath, u.PrintContentWithLineNuber(ymlstr)))

				u.Ppmsgvvvvvhint("yml modified:", modified)

				if localonly {
					f.Vars.Put(reg, modified)
				} else {
					TaskRuntime().ExecbaseVars.Put(reg, modified)
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
						varvalue = Render(varvalueRaw, f.Vars)
					}
				}

				doFlag("localonly", func() {
					localonly = true
				})

				if varname == "" {
					u.InvalidAndExit("validate varname", "the reg varname must not be empty")
				}
				if localonly {
					f.Vars.Put(varname, varvalue)
				} else {
					TaskRuntime().ExecbaseVars.Put(varname, varvalue)
					f.Vars.Put(varname, varvalue)
				}

			})
			u.Ppmsgvvvvvhint("after reg the var - contextual global:", TaskRuntime().ExecbaseVars)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "path_existed":
			cmd := cmdItem.Cmd.(map[interface{}]interface{})
			var raw, path, pathtstr, reg string
			for k, v := range cmd {
				switch k.(string) {
				case "path":
					raw = v.(string)
					path = Render(raw, f.Vars)
					pathtstr = u.Spf("{{.%s}}", path)
				case "reg":
					raw = v.(string)
					reg = Render(raw, f.Vars)
				}
			}

			result := ElementValid(pathtstr, f.Vars)
			TaskRuntime().ExecbaseVars.Put(reg, result)
			f.Vars.Put(reg, result)

		case "return":
			cmdItem.runCmd("array", func() {
				retNames := cmdItem.Cmd.([]interface{})
				var retName string

				if TaskRuntime().ReturnVars == nil {
					TaskRuntime().ReturnVars = core.NewCache()
				}

				for _, v := range retNames {
					rawName := v.(string)
					retName = Render(rawName, f.Vars)
					ret := f.Vars.Get(retName)
					if ret != nil {
						TaskRuntime().ReturnVars.Put(retName, f.Vars.Get(retName))
					} else {
						u.LogWarn("return validation", u.Spf("The referencing var name: (%s) not exist", retName))
					}
				}

			})
			u.Ppmsgvvvvvhint("contextual return vars:", TaskRuntime().ReturnVars)

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
						fromkey = Render(keyRaw, f.Vars)
					}
					if k.(string) == "src" {
						srcRaw := v.(string)
						src = Render(srcRaw, f.Vars)
					}
					if k.(string) == "reg" {
						regRaw := v.(string)
						reg = Render(regRaw, f.Vars)
					}
				}
				doFlag("localonly", func() {
					localonly = true
				})

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
					TaskRuntime().ExecbaseVars.Put(src, reg)
					(*f.Vars).Put(reg, *obj)
				}

			})
			u.Ppmsgvvvvvhint("after reg the var - contextual global:", TaskRuntime().ExecbaseVars)
			u.Ppmsgvvvvvhint("after reg the var - local:", f.Vars)

		case "":
			u.LogWarn("cmd", "temporarily deactivated")

		default:
			u.Pferror("warrning: check cmd name:(%s),%s\n", cmdItem.Name, "cmd not implemented")
		}

	}
}

// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/stephencheng/up/biz/impl"
	u "github.com/stephencheng/up/utils"
	"os"
)

var (
	app = kingpin.New("up", "UP: The Ultimate Provisioner")

	ngo              = app.Command("ngo", "run a entry task")
	ngoTaskName      = ngo.Arg("taskname", "task name to run").Default("Main").String()
	initDefault      = app.Command("init", "create a default skeleton for a quick start")
	list             = app.Command("list", "list tasks")
	listName         = list.Arg("taskname", "task name to inspect").String()
	assist           = app.Command("assist", "assist: templatefunc|")
	assistName       = assist.Arg("assistname", "what to assist").String()
	validate         = app.Command("validate", "validate tasks and plays")
	validateTaskName = validate.Arg("validatetaskname", "taskname").Required().String()
	verbose          = app.Flag("verbose", "verbose level: v-vvvvv").Short('v').String()
	refdir           = app.Flag("refdir", "ref yml files directory").Short('d').String()
	taskfile         = app.Flag("taskfile", "task file to load (without yml extension)").Short('t').String()
	instanceName     = app.Flag("instance", "instance name for execution").Short('i').String()
	configDir        = app.Flag("configdir", "config file directory to load from|default .").String()
	configFile       = app.Flag("configfile", "config file name to load without yml extension|default config").String()
)

func main() {
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	initArgs := func() {
		u.SetConfigYamlDir(*configDir)
		u.SetConfigYamlFile(*configFile)
		u.InitConfig()
		u.ShowCoreConfig()
		u.Pfvvvv(" :release version:  %s", u.CoreConfig.Version)

		u.SetVerbose(*verbose)

		u.SetRefdir(*refdir)
		u.SetTaskfile(*taskfile)
		u.Pfvvvv(" :verbose level:  %s", u.CoreConfig.Verbose)
	}

	switch cmd {
	case initDefault.FullCommand():
		u.Pln("-init default skeleton and configuration")
		impl.InitDefaultSkeleton()

	case ngo.FullCommand():
		initArgs()
		if *ngoTaskName != "" {
			u.Pln("-exec task:", *ngoTaskName)
			t := impl.NewTasker(*instanceName)
			t.ExecTask(*ngoTaskName, nil)
			u.Ptmpdebug("88", u.CoreConfig)
		}

	case list.FullCommand():
		initArgs()
		t := impl.NewTasker(*instanceName)
		if *listName == "=" {
			t.ListAllTasks()
		} else if *listName != "" {
			t.ListTask(*listName)
		} else {
			t.ListTasks()
		}

	case assist.FullCommand():
		u.Pf("-assist: %s\n", *assistName)
		if *assistName == "templatefunc" {
			u.Pln("=List of golang template funcs")
			impl.FuncMapInit()
			impl.ListAllFuncs()
		} else {
			u.LogWarn("What kind of assist do you need?", "Please input a name:")
			u.Pln(`#supported:
templatefunc
`)
		}

	case validate.FullCommand():
		initArgs()
		t := impl.NewTasker(*instanceName)
		taskname := *validateTaskName
		u.Pf("validate task: %s\n")
		t.ValidateTask(taskname)
	}
}


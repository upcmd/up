// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/upcmd/up/biz/impl"
	u "github.com/upcmd/up/utils"
	"os"
)

var (
	app = kingpin.New("up", "UP: The Ultimate Provisioner")

	ngo            = app.Command("ngo", "run an entry task")
	ngoTaskName    = ngo.Arg("taskname", "task name to run").Default("Main").String()
	initDefault    = app.Command("init", "create a default skeleton for a quick start")
	list           = app.Command("list", "list tasks")
	listName       = list.Arg("taskname|=", "task name to inspect").String()
	mod            = app.Command("mod", "module cmd: list | pull | lock | clean | probe")
	modCmd         = mod.Arg("cmd", "list | pull | lock | clean | probe ").Required().String()
	assist         = app.Command("assist", "assist: all_template_func|upcmd_template_func|version")
	assistName     = assist.Arg("assistname", "what to assist").String()
	dryrun         = app.Command("dryrun", "dry run the task")
	dryrunTaskName = dryrun.Arg("dryruntaskname", "taskname").Required().String()
	verbose        = app.Flag("verbose", "verbose level: v-vvvvv").Short('v').String()
	refdir         = app.Flag("refdir", "ref yml files directory").Short('d').String()
	workdir        = app.Flag("workdir", "working directory: cwd | refdir").Short('w').String()
	taskfile       = app.Flag("taskfile", "task file to load (without yml extension)").Short('t').String()
	instanceName   = app.Flag("instance", "instance name for execution").Short('i').String()
	execprofile    = app.Flag("execprofile", "profile name for execution to setup a group environment variables").Short('p').String()
	configDir      = app.Flag("configdir", "config file directory to load from|default .").String()
	configFile     = app.Flag("configfile", "config file name to load without yml extension|default config").String()
)

func main() {
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	initConfig := func() *u.UpConfig {
		cfg := u.NewUpConfig(*configDir, *configFile)
		cfg.SetVerbose(*verbose)
		cfg.SetRefdir(*refdir)
		cfg.SetWorkdir(*workdir)
		cfg.SetTaskfile(*taskfile)
		cfg.InitConfig()
		u.MainConfig = cfg
		cfg.ShowCoreConfig("Main")
		u.Pfvvvv(" :release version:  %s", u.MainConfig.Version)
		u.Pfvvvv(" :verbose level:  %s", u.MainConfig.Verbose)
		u.Pln("work dir:", cfg.AbsWorkDir)
		impl.SetBaseDir(cfg.AbsWorkDir)
		os.Chdir(cfg.AbsWorkDir)
		return cfg
	}()

	impl.FuncMapInit()

	switch cmd {
	case initDefault.FullCommand():
		u.Pln("-init default skeleton and configuration")
		impl.InitDefaultSkeleton()

	case ngo.FullCommand():
		if *ngoTaskName != "" {
			u.Pln("-exec task:", *ngoTaskName)
			if *instanceName != "" && *execprofile != "" {
				u.InvalidAndPanic("parameter validation", "instanceid (-i) and execprofile (-p) can not coexist, please only use one of them")
			}
			t := impl.NewTasker(*instanceName, *execprofile, initConfig)
			impl.Pipein()
			t.ExecTask(*ngoTaskName, nil, false)
		}

	case list.FullCommand():
		t := impl.NewTasker(*instanceName, *execprofile, initConfig)
		if *listName == "=" {
			t.ListAllTasks()
		} else if *listName != "" {
			t.ListTask(*listName)
		} else {
			t.ListTasks()
		}

	case mod.FullCommand():
		t := impl.NewTasker(*instanceName, *execprofile, initConfig)
		if *modCmd == "list" {
			t.ListMainModules()
		}
		if *modCmd == "probe" {
			impl.ListAllModules()
		}
		if *modCmd == "pull" {
			t.PullModules()
		}
		if *modCmd == "lock" {
			t.LockModules()
		}
		if *modCmd == "clean" {
			t.CleanModules()
		}

	case assist.FullCommand():
		u.Pf("-assist: %s\n", *assistName)
		if *assistName == "all_template_func" {
			u.Pln("=List of all template funcs")
			impl.FuncMapInit()
			impl.ListAllFuncs()
		} else if *assistName == "upcmd_template_func" {
			u.Pln("=List of upcmd template funcs")
			impl.FuncMapInit()
			impl.ListUpcmdFuncs()
		} else if *assistName == "version" {
			u.PlnInfo(version_info)
		} else {
			u.LogWarn("What kind of assist do you need?", "Please input a name:")
			u.Pln(`#supported:
all_template_func
upcmd_template_func
version
`)
		}

	case dryrun.FullCommand():
		t := impl.NewTasker(*instanceName, *execprofile, initConfig)
		taskname := *dryrunTaskName
		u.Pf("dryrun task: %s\n")
		t.DryrunTask(taskname)
	}
}

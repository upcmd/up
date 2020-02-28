// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/stephencheng/up/biz/impl"
	"github.com/stephencheng/up/model/core"
	u "github.com/stephencheng/up/utils"
	"os"
)

var (
	app = kingpin.New("up", "UP: The Ultimate Provisioner")

	ngo              = app.Command("ngo", "run a entry task")
	ngoTaskName      = ngo.Arg("taskname", "task name to run").Required().String()
	list             = app.Command("list", "list tasks and plays")
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
	u.SetConfigYamlDir(*configDir)
	u.SetConfigYamlFile(*configFile)
	u.InitConfig()
	u.ShowCoreConfig()
	u.Pfvvvv(" :release version:  %s", u.CoreConfig.Version)

	u.SetVerbose(*verbose)

	u.SetRefdir(*refdir)
	u.SetTaskfile(*taskfile)
	u.Pfvvvv(" :verbose level:  %s", u.CoreConfig.Verbose)
	u.Pfvvvv(" :instance name:  %s", *instanceName)

	core.SetInstanceName(*instanceName)

	switch cmd {
	case ngo.FullCommand():
		if *ngoTaskName != "" {
			u.P("-exec task:", *ngoTaskName)
			impl.InitTasks()
			impl.ExecTask(*ngoTaskName, nil)
		}
	case list.FullCommand():
		impl.InitTasks()
		impl.ListTasks()
	case validate.FullCommand():
		impl.InitTasks()
		taskname := *validateTaskName
		u.Pf("validate task: %s\n")
		impl.ValidateTask(taskname)
	}
}


// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/stephencheng/up/interface/impl"
	u "github.com/stephencheng/up/utils"
	"os"
)

var (
	app = kingpin.New("up", "UP: The Ultimate Provisioner")

	task               = app.Command("task", "run a entry task")
	taskName           = task.Arg("taskname", "task name to run").Required().String()
	list               = app.Command("list", "list tasks and plays")
	listTypeName       = list.Arg("listtypename", "list [ task | flow ]").Required().String()
	validate           = app.Command("validate", "validate tasks and plays")
	validateTypeName   = validate.Arg("validatetypename", "list [ task | flow ]").Required().String()
	validateObjectName = validate.Arg("validateobjectname", "taskname | flowname ]").Required().String()
	play               = app.Command("play", "run a playbook with defined steps")
	playFile           = play.Arg("playfile", "play step file to run").Required().String()
	verbose            = app.Flag("verbose", "verbose level: v-vvvvv").Short('v').String()
	taskdir            = app.Flag("taskdir", "task file directory").Short('d').String()
	taskfile           = app.Flag("taskfile", "task file to load (without yml extension)").Short('t').String()
	instanceName       = app.Flag("instance", "instance name for execution").Short('i').String()
)

func main() {

	u.InitConfig()
	u.ShowCoreConfig()
	u.Pfvvvv(" :release version:  %s", u.CoreConfig.Version)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	u.SetVerbose(*verbose)

	u.SetTaskdir(*taskdir)
	u.SetTaskfile(*taskfile)
	u.Pfvvvv(" :verbose level:  %s", u.CoreConfig.Verbose)
	u.Pfvvvv(" :instance name:  %s", *instanceName)
	impl.SetInstanceName(*instanceName)

	switch cmd {
	case task.FullCommand():
		if *taskName != "" {
			u.P("-exec task:", *taskName)
			impl.InitTasks()
			impl.ExecTask(*taskName, nil)
		}
	case list.FullCommand():
		u.P("-list", *listTypeName)
		switch *listTypeName {
		case "task":
			impl.InitTasks()
			impl.ListTasks()
		case "flow":
		}
	case validate.FullCommand():
		u.P("-validate", *validateTypeName)
		switch *validateTypeName {
		case "task":
			impl.InitTasks()
			taskname := *validateObjectName
			u.Pf("validate task: %s\n")
			impl.ValidateTask(taskname)
		case "flow":
		}
	case play.FullCommand():
		u.P(*playFile)
	}
}


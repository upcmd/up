// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/alecthomas/kingpin"
	svc "github.com/stephencheng/up/service"
	u "github.com/stephencheng/up/utils"
	"os"
)

var (
	app = kingpin.New("up", "UP: The Ultimate Provisioner")

	task         = app.Command("task", "run a entry task")
	taskName     = task.Arg("taskname", "task name to run").Required().String()
	list         = app.Command("list", "list tasks and plays")
	listTypeName = list.Arg("typename", "list [ task | flow ]").Required().String()
	play         = app.Command("play", "run a playbook with defined steps")
	playFile     = play.Arg("playfile", "play step file to run").Required().String()
	verbose      = app.Flag("verbose", "verbose level: v-vvvvv").Short('v').String()
	taskdir      = app.Flag("taskdir", "task file directory").Short('d').String()
	taskfile     = app.Flag("taskfile", "task file to load (without yml extension)").Short('t').String()
)

func main() {

	u.InitConfig()
	u.ShowCoreConfig()
	u.Pfvvvv(" :release version:", u.CoreConfig.Version)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	u.SetVerbose(*verbose)

	u.SetTaskdir(*taskdir)
	u.SetTaskfile(*taskfile)
	u.Pfvvvv(" :verbose level:%s", u.CoreConfig.Verbose)

	switch cmd {
	case task.FullCommand():
		if *taskName != "" {
			u.P("-exec task:", *taskName)
			svc.InitTasks()
			svc.ExecTask(*taskName)
		}
	case list.FullCommand():
		u.P("-list", *listTypeName)
		switch *listTypeName {
		case "task":
			svc.InitTasks()
			svc.ListTasks()
		case "flow":
		}
	case play.FullCommand():
		u.P(*playFile)

	}
}


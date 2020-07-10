// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"bufio"
	u "github.com/upcmd/up/utils"
	"io"
	"os"
)

func Pipein() {
	info, err := os.Stdin.Stat()
	if err != nil {
		u.LogErrorAndExit("Pipe in error", err, "please double check you CLI pipe in syntax")
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	var pipeinchars []rune

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		pipeinchars = append(pipeinchars, input)
	}

	pipeinstr := string(pipeinchars)
	UpRunTimeVars.Put(UP_RUNTIME_TASK_PIPE_IN_CONTENT, pipeinstr)
}

// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/term"
	"io"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
	"os"
	"strings"
)

var command = flag.String("c", "", "command to be executed")

//var command = flag.String("c", "echo hello", "command to be executed")

func main_test() {
	//Run(".", "ls")
	//Run(".", "git status")
	//Run("/tmp", "ls -lartG")
	//Run("", "echo PATH=$PATH")
}

func main_run_args() {
	os.Chdir("/tmp")
	flag.Parse()
	err := runAll()
	if e, ok := interp.IsExitStatus(err); ok {
		os.Exit(int(e))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RunShellCmd(dir string, command string) error {
	cwd, err := os.Getwd()
	if dir != "" {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			os.Chdir(dir)
		} else {
			LogErrorAndExit("check dir existence", err, "exec path does not exist")
		}
	}

	r, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		fmt.Println("error: init terminal errored", err)
	}

	if command != "" {
		err = run(r, strings.NewReader(command), "")
	}
	os.Chdir(cwd)

	//if e, ok := interp.IsExitStatus(err); ok {
	//	LogErrorAndContinue("shell exec failed", e, "please exam the error and fix the problem and retry again")
	//}

	if err != nil {
		LogErrorAndContinue("shell exec failed", err, "please exam the error and fix the problem and retry again")
	}

	return err

}

func runAll() error {
	r, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		return err
	}

	if *command != "" {
		return run(r, strings.NewReader(*command), "")
	}
	if flag.NArg() == 0 {
		if term.IsTerminal(int(os.Stdin.Fd())) {
			return runInteractive(r, os.Stdin, os.Stdout, os.Stderr)
		}
		return run(r, os.Stdin, "")
	}
	for _, path := range flag.Args() {
		fmt.Println(44, path)
		if err := runPath(r, path); err != nil {
			return err
		}
	}
	return nil
}

func run(r *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	r.Reset()
	ctx := context.Background()
	return r.Run(ctx, prog)
}

func runPath(r *interp.Runner, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return run(r, f, path)
}

func runInteractive(r *interp.Runner, stdin io.Reader, stdout, stderr io.Writer) error {
	parser := syntax.NewParser()
	fmt.Fprintf(stdout, "$ ")
	var runErr error
	fn := func(stmts []*syntax.Stmt) bool {
		if parser.Incomplete() {
			fmt.Fprintf(stdout, "> ")
			return true
		}
		ctx := context.Background()
		for _, stmt := range stmts {
			runErr = r.Run(ctx, stmt)
			if r.Exited() {
				return false
			}
		}
		fmt.Fprintf(stdout, "$ ")
		return true
	}
	if err := parser.Interactive(stdin, fn); err != nil {
		return err
	}
	return runErr
}


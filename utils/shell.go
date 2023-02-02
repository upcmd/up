package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
	"os"
	"strings"
)

type ExecResult struct {
	Cmd    string
	Code   int
	Output string
	ErrMsg string
}

type CmdOpts struct {
	Command string
	Dir     string
	Env     []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

func getEnvs(envs *map[string]string) []string {
	environ := os.Environ()
	for k, v := range *envs {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}
	return environ
}

func RunCmd(cmd string, dir string, envs *map[string]string) ExecResult {
	stdOut := bytes.NewBufferString("")
	stdErr := bytes.NewBufferString("")
	stdin := os.Stdin

	err := runCmdWithOps(&CmdOpts{
		Command: cmd,
		Dir:     dir,
		Env:     getEnvs(envs),
		Stdin:   stdin,
		Stdout:  stdOut,
		Stderr:  stdErr,
	})
	var result ExecResult
	var errored bool = false
	if statusCode, ok := interp.IsExitStatus(err); ok {
		errored = true
		result.Code = int(statusCode)
		result.ErrMsg = stdErr.String()
	}

	if !errored {
		result.Code = 0
		result.Output = stdOut.String()
	}

	return result
}

func runCmdWithOps(opts *CmdOpts) error {
	p, err := syntax.NewParser().Parse(strings.NewReader(opts.Command), "")
	if err != nil {
		return err
	}

	environ := opts.Env
	if len(environ) == 0 {
		environ = os.Environ()
	}

	r, err := interp.New(
		interp.Dir(opts.Dir),
		interp.Env(expand.ListEnviron(environ...)),
		interp.StdIO(opts.Stdin, opts.Stdout, opts.Stderr),
	)
	if err != nil {
		return err
	}
	return r.Run(context.Background(), p)
}

func RunSimpleCmd(dir string, command string) error {
	if dir != "" {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
		} else {
			LogErrorAndPanic("check dir existence", err, "exec path does not exist")
		}
	}

	r, err := interp.New(
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
		interp.Dir(dir),
	)
	if err != nil {
		fmt.Println("error: init terminal errored", err)
	}

	if command != "" {
		err = runSimple(r, strings.NewReader(command), "")
	}

	if err != nil {
		LogErrorAndContinue("shell exec failed", err, "please exam the error and fix the problem and retry again")
	}

	return err

}

func runSimple(r *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	r.Reset()
	ctx := context.Background()
	return r.Run(ctx, prog)
}

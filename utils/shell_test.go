package utils

import (
	"testing"
)

func TestRunShellCmd(t *testing.T) {
	RunSimpleCmd("", "export CM_TEST_XX=haha && env |grep CM_TEST_")
	RunSimpleCmd("", "pwd")
	RunSimpleCmd("/tmpxx", "pwd")
}

func TestRunShellCmdWithEnvVars01(t *testing.T) {
	result := RunCmd("pwd",
		"/tmp",
		&map[string]string{
			"CM_TEST1": "CM_TEST1_value",
			"CM_TEST2": "CM_TEST2_value",
		},
	)
	Pln(result.Code)
	Pln(result.Output)
	Pln(result.ErrMsg)
}

func TestRunShellCmdWithEnvVars02(t *testing.T) {
	result := RunCmd("ls -l /dir_not_exist",
		"/tmp",
		&map[string]string{
			"CM_TEST1": "CM_TEST1_value",
			"CM_TEST2": "CM_TEST2_value",
		},
	)
	Pln(result.Code)
	Pln(result.Output)
	Pln(result.ErrMsg)
}

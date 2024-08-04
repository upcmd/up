package functests

import (
	"github.com/upcmd/up/biz/impl"
	"github.com/upcmd/up/tests"
	u "github.com/upcmd/up/utils"
	"path"
	"testing"
)

func TestC(t *testing.T) {
	dirs := tests.GetModuleTestCollection()

	for _, x := range dirs {
		u.Pln("==testing:", x, "==")
		cfg := tests.SetupMx(path.Join(x))
		t := impl.NewTasker("dev", "", cfg)
		t.ExecTask("Main", nil, false)
		t.Unset()
	}
}

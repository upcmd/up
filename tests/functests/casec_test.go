package functests

import (
	"github.com/upcmd/up/biz/impl"
	"github.com/upcmd/up/tests"
	u "github.com/upcmd/up/utils"
	"os"
	"testing"
)

func init() {
	os.Chdir("../..")
}

func TestC(t *testing.T) {
	cfg := u.NewUpConfig("", "").InitConfig()
	u.MainConfig = cfg
	files := tests.GetUnitTestCollection()
	impl.FuncMapInit()

	for _, x := range files {
		u.Pln("testing:", x)
		u.Pln("work dir:", cfg.AbsWorkDir)
		tests.Setupx(x, cfg)
		t := impl.NewTasker("dev", "", cfg)
		t.ExecTask("task", nil, false)
		t.Unset()
	}
}

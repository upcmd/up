package tests

import (
	"github.com/upcmd/up/biz/impl"
	u "github.com/upcmd/up/utils"
	"os"
)

//mock required settings
func SetupMx(dirpath string) *u.UpConfig {
	cfg := u.NewUpConfig(dirpath, "")
	cfg.Secure = &u.SecureSetting{Type: "default_aes", Key: "enc_key"}
	cfg.RefDir = dirpath
	cfg.WorkDir = "refdir"
	cfg.InitConfig()
	u.MainConfig = cfg
	wkdir := cfg.AbsWorkDir
	u.Pln("work dir:", wkdir)
	impl.SetBaseDir(wkdir)
	os.Chdir(wkdir)
	cfg.ShowCoreConfig("moduletest")
	u.Ppmsgvvvvhint("core config", cfg)
	u.Pln(" :test task file:", cfg.TaskFile)
	u.Pln(" :release version:", cfg.Version)
	u.Pln(" :verbose level:", cfg.Verbose)
	return cfg
}

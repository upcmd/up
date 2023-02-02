package tests

import (
	"fmt"
	"github.com/spf13/viper"
	//"github.com/upcmd/up/model"
	"os"
	"path"
	"runtime"
)

var (
	ymlroot              *viper.Viper
	RootDir, TestDataDir = getDirs()
)

func getDirs() (string, string) {
	_, filename, _, _ := runtime.Caller(1)
	utilsDir := path.Dir(filename)
	rootDir := path.Join(utilsDir, "..")
	return rootDir, path.Join(rootDir, "./testdata/poc")
}

func MockLoadYml(path, filename string) *viper.Viper {

	viper.AddConfigPath(".") //test in current directory as top priority then the configured path
	viper.AddConfigPath(path)
	viper.SetConfigType("yaml")
	viper.SetConfigName(filename)

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("yml file: %s/%s.yml not found...", path, filename)
		fmt.Println("errored:", err.Error())
		os.Exit(3)
	}

	ymlroot = viper.GetViper()
	return viper.GetViper()

}

//func GetNode(node string) model.Step {
//
//	var step model.Step
//	cfgEntry := ymlroot.Sub(node)
//
//	err := cfgEntry.Unmarshal(&step)
//
//	if err != nil {
//		fmt.Println("unable to decode into struct:", err.Error())
//	}
//	return step
//}

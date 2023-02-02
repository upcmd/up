package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func YamlLoader(id, path, filename string) *viper.Viper {
	Pf("loading [%s]:  %s/%s\n", id, path, filename)
	newV := viper.New()
	newV.AddConfigPath(path)
	newV.SetConfigType("yaml")
	newV.SetConfigName(filename)

	err := newV.ReadInConfig()

	if err != nil {
		fmt.Printf("yml dir: %s\n", path)
		fmt.Printf("yml file: %s\n", filename)
		fmt.Println("errored:", err.Error())
		LogError("Yaml loading error", err)
		DebugYmlContent(path, filename)
		os.Exit(3)
	}

	return newV
}

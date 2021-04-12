package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var RootPath = ""

func GetProjectPath() string {
	var projectPath string
	projectPath, _ = os.Getwd()
	return projectPath
}

func init() {
	RootPath = os.Getenv("RootPath")
	if RootPath == "" {
		RootPath = filepath.Join(GetProjectPath())
		RootPath = "/Users/hxx/Github/hmq"
	}
	fmt.Println(">>> RootPath:", RootPath)

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(RootPath) // call multiple times to add many search paths
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic("init| config failure!!!")
	}
}

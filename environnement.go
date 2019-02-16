package main

import (
	"github.com/spf13/viper"
	"log"
	_ "os"
	_ "path/filepath"
)

const dir_import = "db/images/import/"
const dir_original = "db/images/original/"

var Config *viper.Viper

func init() {
	Config = viper.New()
}

func LoadDefaultConf() {
	Config.SetConfigName("config")
	Config.AddConfigPath(".")
	err := Config.ReadInConfig()
	if err != nil {
		log.Panic(err)
	}
}

func DirFileImport() string {
	return dir_import
}

func DirFileStorage() string {
	return dir_original
}

func BaseDir() string {
	/*
		ex, errEx := os.Executable()
		if errEx != nil {
			log.Panic(errEx)
		}
		return filepath.Dir(ex) + "/"
	*/
	return "./"
}

func DatabasePath() string {
	path := Config.GetString("database-path")
	return BaseDir() + path
}

func ConfigValue(AParamName string) string {
	value := Config.GetString(AParamName)
	return value
}

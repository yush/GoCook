package main

import (
	"github.com/olebedev/config"
	"log"
	"os"
	"path/filepath"
)

const dir_import = "db/images/import/"
const dir_original = "db/images/original/"

var Conf *config.Config
var globalConf *config.Config

func init() {
	globalConf, errYml := config.ParseYamlFile(BaseDir() + "config.yml")
	if errYml != nil {
		log.Panic(errYml)
	}

	var err error
	Conf, err = globalConf.Get("development")
	if err != nil {
		log.Panic(err)
	}
}

func SetTestMode() {
	var err error
	Conf, err = globalConf.Get("test")
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
	ex, errEx := os.Executable()
	if errEx != nil {
		log.Panic(errEx)
	}
	return filepath.Dir(ex) + "/"
}

func DatabasePath() string {
	path, err := Conf.String("database-path")
	if err != nil {
		log.Panic(err)
	}
	return BaseDir() + path
}

func ConfigValue(AParamName string) string {
	value, err := Conf.String(AParamName)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return value
}

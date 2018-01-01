package main

import (
	"github.com/olebedev/config"
	"log"
)

const dir_import = "db/images/import/"
const dir_original = "db/images/original/"

var Conf *config.Config
var globalConf *config.Config

func init() {
	var err error
	if globalConf, err = config.ParseYamlFile("config.yml"); err != nil {
		log.Panic(err)
	}

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
	dir, err := Conf.String("root")
	if err != nil {
		panic("file not found")
	}
	return dir
}

func DatabasePath() string {
	path, err := Conf.String("database-path")
	if err != nil {
		log.Panic(err)
	}
	return BaseDir() + path
}

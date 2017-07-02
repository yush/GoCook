package main

import (
	"os"
)

const dir_import = "db/images/import/"
const dir_original = "db/images/original/"

func init() {

}

func DirFileImport() string {
	return dir_import
}

func DirFileStorage() string {
	return dir_original
}

func BaseDir() string {
	return os.Getenv("GOCOOK_BASEDIR")
}

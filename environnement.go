package main

import ()

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
	var dir string
	err := Conf.Get("dir", &dir)
	if err != nil {
		panic("file not found")
	}
	return dir
}

package main

import (
	"database/sql"
	"github.com/jlaffaye/ftp"
	_ "github.com/jlaffaye/ftp"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func BackupDb(backupFileName string) error {
	os.Remove(backupFileName)

	srcDb, err := sql.Open("sqlite3_backup", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Println(err)
	}
	defer srcDb.Close()
	srcDb.Ping()
	srcConn := Sqlite3conn[len(Sqlite3conn)-1]

	destDb, err := sql.Open("sqlite3_backup", backupFileName)
	if err != nil {
		log.Println(err)
	}
	defer destDb.Close()
	destDb.Ping()
	destConn := Sqlite3conn[len(Sqlite3conn)-1]

	if err != nil {
		log.Println(err)
	}
	bk, err := destConn.Backup("main", srcConn, "main")
	if err != nil {
		log.Println(err)
	}

	// single step backup
	bk.Step(-1)
	if err != nil {
		log.Println(err)
	}
	_, err = destDb.Query("SELECT * FROM recipes ")
	if err != nil {
		log.Println(err)
	}
	bk.Finish()
	return nil
}

func BackupFiles(test string) {
}

func getBackupDirName() string {
	t := time.Now()
	return ConfigValue("backup-ftp-folder") + t.Format("2006-01-02-15-04-05")
}

func BackupToFTP() {
	backupFileName := "gocook.backup.db3"
	backupName := BaseDir() + "/db/" + backupFileName
	backupDirName := getBackupDirName()
	BackupDb(backupName)
	r, err := os.Open(backupName)
	if err != nil {
		log.Println(err)
	}

	c, err := connectToBackupServer()
	err = c.MakeDir(backupDirName)
	if err != nil {
		log.Println(err)
	}
	c.ChangeDir(backupDirName)

	c.Stor(backupFileName, r)

	files, err := ioutil.ReadDir(BaseDir() + DirFileStorage())
	for i := range files {
		r, err := os.Open(BaseDir() + DirFileStorage() + files[i].Name())
		if err != nil {
			println(err)
		}
		err = c.Stor(files[i].Name(), r)
		if err != nil {
			println(err)
		}
	}

	err = c.Quit()
	if err != nil {
		log.Println(err)
	}
}

func connectToBackupServer() (*ftp.ServerConn, error) {
	c, err := ftp.DialTimeout(ConfigValue("backup-ftp-server")+":21", 5*time.Second)
	if err != nil {
		log.Println(err)
	}
	err = c.Login(ConfigValue("backup-ftp-login"), ConfigValue("backup-ftp-password"))
	if err != nil {
		log.Println(err)
	}
	return c, err
}

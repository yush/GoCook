package main

import (
	"database/sql"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
)

type Recipe struct {
	Id       uint
	Name     string
	Filepath string
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllRecipes(db *sql.DB) []Recipe {
	recipes := make([]Recipe, 0, 10)
	db.Begin()
	rows, err := db.Query("select id, name, filepath from recipes")

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.Id, &r.Name, &r.Filepath); err != nil {
			log.Fatal(err)
		} else {
			recipes = append(recipes, r)
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return recipes
}

func Insert(db *sql.DB, name string, filename string) {
	// INSERT
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into RECIPES(NAME, FILEPATH) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, filename)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func GetRecipe(db *sql.DB, id int) Recipe {
	var recipe Recipe
	err := db.QueryRow("select ID, NAME, FILEPATH from RECIPES where ID = :id", id).Scan(&recipe.Id, &recipe.Name, &recipe.Filepath)
	if err != nil {
		log.Fatal(err)
	}
	return recipe
}

func UploadRecipe(db *sql.DB, img image.Image, handler *multipart.FileHeader, recipeName string) {
	resizeAndAddFile(handler.Filename, img)
	Insert(db, recipeName, handler.Filename)
}

func ImportRecipes(db *sql.DB, dirname string) {
	existingFiles, err := ioutil.ReadDir(dirname)
	checkErr(err)

	files, err := ioutil.ReadDir(BaseDir() + DirFileImport())
	checkErr(err)
	for _, fileInfo := range files {
		IsExisting := false
		for i := range existingFiles {
			if existingFiles[i].Name() == fileInfo.Name() {
				IsExisting = true
				break
			}
		}

		if IsExisting == false {
			Insert(db, fileInfo.Name(), fileInfo.Name())
			// Read all content of src to data
			data, errLoad := ioutil.ReadFile(fileInfo.Name())
			checkErr(errLoad)
			err := addFile(fileInfo.Name(), data)
			checkErr(err)
		}
		os.Remove(BaseDir() + DirFileStorage() + fileInfo.Name())
	}
}
func addFile(filename string, data []byte) error {
	// Write data to dst
	err := ioutil.WriteFile(BaseDir()+DirFileStorage()+filename, data, 0644)
	return err
}

func resizeAndAddFile(name string, img image.Image) error {
	/*
		var out image.Image;
		errConvert := rez.Convert(out, img, rez.NewBicubicFilter())
		if errConvert!= nil {
			return errConvert
		}
	*/
	out := resize.Resize(1024, 768, img, resize.Lanczos3)

	toimg, _ := os.Create(BaseDir() + DirFileStorage() + name)
	defer toimg.Close()
	errEncode := jpeg.Encode(toimg, out, &jpeg.Options{jpeg.DefaultQuality})
	if errEncode != nil {
		return errEncode
	}
	return nil
}

func RemoveFile(filename string) {
	err := os.Remove(BaseDir() + DirFileStorage() + filename)
	if err != nil {
		log.Print(err)
	}
}

func DeleteRecipe(db *sql.DB, id int) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("DELETE FROM RECIPES where ID=?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

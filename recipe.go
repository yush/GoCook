package main

import (
	"database/sql"
	"fmt"
	"io"
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

func GetRecipe(db *sql.DB, id string) Recipe {
	var recipe Recipe
	err := db.QueryRow("select ID, NAME, FILEPATH from RECIPES where ID = :id", id).Scan(&recipe.Id, &recipe.Name, &recipe.Filepath)
	if err != nil {
		log.Fatal(err)
	}
	return recipe
}

func ImportRecipes(db *sql.DB, dirname string) {

	existingFiles, err := ioutil.ReadDir(dirname)
	checkErr(err)

	files, err := ioutil.ReadDir(BaseDir() + DirFileImport())
	checkErr(err)
	for _, file := range files {
		IsExisting := false
		for i := range existingFiles {
			if existingFiles[i].Name() == file.Name() {
				IsExisting = true
				break
			}
		}

		if IsExisting == false {

			Insert(db, file.Name(), file.Name())
			// Read all content of src to data
			data, err := ioutil.ReadFile(BaseDir() + DirFileImport() + file.Name())
			checkErr(err)
			// Write data to dst
			err = ioutil.WriteFile(BaseDir()+DirFileStorage()+file.Name(), data, 0644)
			checkErr(err)
		}
		os.Remove(BaseDir() + DirFileStorage() + file.Name())
	}
}

func UploadRecipe(db *sql.DB, file multipart.File, handler *multipart.FileHeader, recipeName string) {
	f, err := os.Create(DirFileStorage() + recipeName)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	_, err = io.Copy(f, file) // copy the image

	if err != nil {
		log.Fatal("Something was wrong")
	}
	Insert(db, recipeName, handler.Filename)
}

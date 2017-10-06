package main

import (
	"database/sql"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path"
	"strings"
)

type Recipe struct {
	Id       uint
	Name     string
	Filepath string
}

type Category struct {
	Id      uint
	Name    string
	User_id uint
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllRecipes(db *sql.DB, userId int) []Recipe {
	recipes := make([]Recipe, 0, 10)
	db.Begin()
	rows, err := db.Query("select id, name, filepath from recipes where OWNERID = ?", userId)

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

func Insert(db *sql.DB, userId int, name string, filename string) {
	// INSERT
	var newId int
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	db.QueryRow("SELECT MAX(ID) FROM RECIPES").Scan(&newId)
	stmt, err := tx.Prepare("insert into RECIPES(ID, NAME, FILEPATH, OWNERID) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newId+1, name, filename, userId)
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

func UploadRecipe(db *sql.DB, userId int, img image.Image, handler *multipart.FileHeader, recipeName string) {
	resizeAndAddFile(handler.Filename, img)
	Insert(db, userId, recipeName, handler.Filename)
}

func ImportRecipes(db *sql.DB, userId int, dirname string) {
	var IsSupportedFormat bool
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

		ext := strings.ToLower(path.Ext(fileInfo.Name()))
		switch ext {
		case ".png", ".jpg", ".gif":
			IsSupportedFormat = true
		default:
			IsSupportedFormat = false
		}

		if IsExisting {
			log.Print("Unable to import ?: already exists", fileInfo.Name())
			continue
		}

		if !IsSupportedFormat {
			log.Print("Unable to import ?: format not supported ", fileInfo.Name())
			continue
		}

		Insert(db, userId, fileInfo.Name(), fileInfo.Name())
		// Read all content of src to data
		data, errLoad := ioutil.ReadFile(BaseDir() + DirFileImport() + fileInfo.Name())
		checkErr(errLoad)
		err := addFile(fileInfo.Name(), data)
		checkErr(err)
		os.Remove(BaseDir() + DirFileImport() + fileInfo.Name())
	}
}
func addFile(filename string, data []byte) error {
	// Write data to dst
	newFile := BaseDir() + DirFileStorage() + filename
	err := ioutil.WriteFile(newFile, data, 0644)
	return err
}

func resizeAndAddFile(name string, img image.Image) error {
	out := resizeFile(img, 960)

	toimg, _ := os.Create(BaseDir() + DirFileStorage() + name)
	defer toimg.Close()
	errEncode := jpeg.Encode(toimg, out, &jpeg.Options{jpeg.DefaultQuality})
	if errEncode != nil {
		return errEncode
	}
	return nil
}

func resizeFile(img image.Image, max int) image.Image {
	var out image.Image
	var height uint
	var width uint
	size := img.Bounds().Size()
	max_img := math.Max(float64(size.X), float64(size.Y))
	if max_img > float64(max) {
		scale := max_img / float64(max)
		height = uint(float64(size.Y) / scale)
		width = uint(float64(size.X) / scale)
		out = resize.Resize(width, height, img, resize.Lanczos3)
	} else {
		out = img
	}
	return out
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

func NewCategory(db *sql.DB, name string, user_id int) {
	var newId int
	db.Begin()
	defer db.Close()

	db.QueryRow("SELECT MAX(ID) FROM CATEGORIES").Scan(&newId)
	_, err := db.Exec("INSERT INTO CATEGORIES(ID, NAME, USER_ID) values (?, ?, ?)", newId+1, name, user_id)
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllCategories(db *sql.DB, user_id int) []Category {
	db.Begin()
	defer db.Close()

	rows, err := db.Query("SELECT ID, NAME FROM CATEGORIES WHERE USER_ID = ?", user_id)
	if err != nil {
		log.Fatal(err)
	}
	categories := make([]Category, 0, 10)
	for rows.Next() {
		var cat Category
		rows.Scan(&cat.Id, &cat.Name)
		categories = append(categories, cat)
	}
	return categories
}

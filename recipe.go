package main

import (
	"bufio"
	"database/sql"
	"image"
	_ "image/gif"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"github.com/nfnt/resize"
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

type ByName []Recipe

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return strings.ToLower(a[i].Name) < strings.ToLower(a[j].Name) }

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func retrieveRecipesFromQuery(rows *sql.Rows) []Recipe {
	recipes := make([]Recipe, 0, 10)

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

func GetAllRecipes(db *sql.DB, userId uint) []Recipe {

	db.Begin()
	rows, err := db.Query("select recipes.id, name, filepath from recipes join"+
		" CATEGORIES_DETAILS on recipes.ID = RECIPE_ID "+
		"WHERE CATEGORIES_DETAILS.USER_ID = ?", userId)
	if err != nil {
		log.Fatal(err)
	}

	return retrieveRecipesFromQuery(rows)
}

func GetAllRecipesForCat(db *sql.DB, userId uint, CatId uint) []Recipe {
	db.Begin()
	rows, err := db.Query("select recipes.id, name, filepath from recipes join"+
		" CATEGORIES_DETAILS on recipes.ID = RECIPE_ID "+
		"WHERE CATEGORIES_DETAILS.USER_ID = ? AND CATEGORY_ID = ?", userId, CatId)

	if err != nil {
		log.Fatal(err)
	}

	return retrieveRecipesFromQuery(rows)
}

func Insert(db *sql.DB, userId uint, name string, filename string) uint {
	// INSERT
	var newId uint
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	db.QueryRow("SELECT MAX(ID) FROM RECIPES").Scan(&newId)
	newId = newId + 1
	stmt, err := tx.Prepare("insert into RECIPES(ID, NAME, FILEPATH, OWNERID) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newId, name, filename, userId)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
	return newId
}

func GetRecipe(db *sql.DB, id int) Recipe {
	var recipe Recipe
	err := db.QueryRow("select ID, NAME, FILEPATH from RECIPES where ID = :id", id).Scan(&recipe.Id, &recipe.Name, &recipe.Filepath)
	if err != nil {
		log.Fatal(err)
	}
	return recipe
}

func UpdateRecipe(db *sql.DB, recipe Recipe) error {
	_, err := db.Exec("UPDATE RECIPES SET NAME= ? WHERE ID = ?", recipe.Name, recipe.Id)
	return err

}

func UploadRecipe(db *sql.DB, userId uint, img image.Image, fileName string, recipeName string) int {
	resizeAndAddFile(fileName, img)
	newID := Insert(db, userId, recipeName, fileName)

	err := InsertCategoryDetails(db, userId, newID, 0)
	if err != nil {
		println(err)
	}
	return int(newID)
}

func ImportRecipes(db *sql.DB, userId uint, dirname string) error {
	var IsSupportedFormat bool
	DirImport := BaseDir() + DirFileImport()
	existingFiles, err := ioutil.ReadDir(dirname)
	checkErr(err)

	files, err := ioutil.ReadDir(DirImport)
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

		f, err := os.Open(DirImport + fileInfo.Name())
		if err != nil {
			return err
		}
		defer f.Close()

		img, _, errDecode := image.Decode(bufio.NewReader(f))
		if errDecode == nil {
			recipeName := getRecipeNameFromFileName(fileInfo.Name())
			UploadRecipe(db, userId, img, fileInfo.Name(), recipeName)
			os.Remove(BaseDir() + DirFileImport() + fileInfo.Name())
		} else {
			println(errDecode)
		}
	}
	return nil
}

func getRecipeNameFromFileName(FileName string) string {
	return strings.TrimSuffix(FileName, path.Ext(FileName))
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
	errEncode := jpeg.Encode(toimg, out, &jpeg.Options{Quality: jpeg.DefaultQuality})
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

func DeleteRecipe(db *sql.DB, userId uint, recipeId uint) error {
	Failed := false
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmtRecipe, err := tx.Prepare("DELETE FROM RECIPES where ID=?")
	if err != nil {
		return err
	}
	defer stmtRecipe.Close()

	stmtCatDetails, err := tx.Prepare("DELETE FROM CATEGORIES_DETAILS WHERE USER_ID = ? AND RECIPE_ID = ?")

	_, err = stmtRecipe.Exec(recipeId)
	if err != nil {
		Failed = true
	}

	_, err = stmtCatDetails.Exec(userId, recipeId)
	if err != nil {
		Failed = true
	}

	if !Failed {
		tx.Commit()
	} else {
		tx.Rollback()
		return err
	}
	return nil
}

func NewCategory(db *sql.DB, name string, user_id uint) {
	var newId int
	db.Begin()
	defer db.Close()

	db.QueryRow("SELECT MAX(ID) FROM CATEGORIES").Scan(&newId)
	_, err := db.Exec("INSERT INTO CATEGORIES(ID, NAME, USER_ID) values (?, ?, ?)", newId+1, name, user_id)
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllCategories(db *sql.DB, user_id uint) []Category {
	db.Begin()
	defer db.Close()

	rows, err := db.Query("SELECT ID, NAME FROM CATEGORIES WHERE USER_ID = ?", user_id)
	if err != nil {
		log.Fatal(err)
	}
	categories := make([]Category, 0, 10)
	categories = append(categories, Category{Id: 0, Name: "No category", User_id: user_id})
	for rows.Next() {
		var cat Category
		rows.Scan(&cat.Id, &cat.Name)
		categories = append(categories, cat)
	}
	return categories
}

func InsertCategoryDetails(db *sql.DB, UserId uint, RecipeId uint, NewCatId int) error {
	var newId uint

	db.QueryRow("SELECT MAX(ID) FROM CATEGORIES_DETAILS").Scan(&newId)
	_, err := db.Exec("INSERT INTO CATEGORIES_DETAILS(ID, CATEGORY_ID, RECIPE_ID, USER_ID) VALUES (?, ?, ?, ?)", newId+1, NewCatId, RecipeId, UserId)
	return err
}

func UpdateCategoryDetails(db *sql.DB, UserId uint, RecipeId int, NewCatId int) error {
	_, err := db.Exec("UPDATE CATEGORIES_DETAILS SET CATEGORY_ID= ? WHERE USER_ID = ? AND RECIPE_ID = ?", NewCatId, UserId, RecipeId)
	return err
}

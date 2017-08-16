package main

import (
	"database/sql"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"image"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var templates map[string]*template.Template

func init() {
	loadTemplates()
}

func main() {

	router := httprouter.New()
	router.ServeFiles("/public/*filepath", http.Dir(BaseDir()+"public/"))
	router.ServeFiles("/images/*filepath", http.Dir(BaseDir()+"db/images/"))
	router.GET("/signin", SigninRoute)
	router.GET("/signup", SignupRoute)
	router.GET("/categories", ListCategories)
	router.GET("/newcategory", NewCategories)
	router.POST("/categories", PostNewCategories)
	router.GET("/recipes", RecipesRoute)
	router.POST("/deleterecipes", DeleteRecipesRoute)
	router.GET("/newrecipe", GetNewRecipeHandler)
	router.GET("/recipes/:id", GetRecipeHandler)
	router.POST("/recipes", PutRecipeHandler)
	router.GET("/import", ImportRoute)
	router.GET("/", IndexRoute)

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func getDb() *sql.DB {
	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func IndexRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if err := templates["index"].Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func SigninRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := templates["signin"].Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func SignupRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := templates["signup"].Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func NewCategories(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := templates["newCategory"].Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func ListCategories(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db := getDb()
	defer db.Close()

	user_id := 1
	categories := GetAllCategories(db, user_id)
	if err := templates["listCategories"].Execute(res, categories); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func PostNewCategories(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db := getDb()
	defer db.Close()

	NewCategory(db, req.FormValue("name"), 1)
	http.Redirect(res, req, "/categories", 301)
}

func RecipesRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db := getDb()
	defer db.Close()

	recipes := GetAllRecipes(db)
	if err := templates["recipes"].Execute(res, recipes); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func DeleteRecipesRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db := getDb()
	defer db.Close()

	req.ParseForm()
	for i := range req.PostForm {
		if strings.Contains(i, "check-") {
			strIdRecipe := strings.TrimPrefix(i, "check-")
			idRecipe, err := strconv.Atoi(strIdRecipe)
			if err != nil {
				log.Fatal(err)
			}
			recipe := GetRecipe(db, idRecipe)
			RemoveFile(recipe.Filepath)
			DeleteRecipe(db, idRecipe)
		}
	}
	http.Redirect(res, req, "/recipes", 301)
}

func GetRecipeHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	db := getDb()
	defer db.Close()

	idRecipe, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		log.Fatal(err)
	}

	recipe := GetRecipe(db, idRecipe)
	if err := templates["showRecipe"].Execute(res, recipe); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func GetNewRecipeHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := templates["newRecipe"].Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func PutRecipeHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	db := getDb()
	defer db.Close()

	file, handler, err := req.FormFile("uploadfile")
	if err != nil {
		log.Fatal(err)
	}

	img, _, errDecode := image.Decode(file)
	if errDecode != nil {
		log.Fatal(errDecode)
	}
	UploadRecipe(db, img, handler, req.FormValue("name"))
	http.Redirect(res, req, "/recipes", 301)
}

func ImportRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db := getDb()
	defer db.Close()

	ImportRecipes(db, BaseDir()+DirFileStorage())
	http.Redirect(res, req, "/recipes", 301)
}

func loadTemplates() {
	var baseTemplate = BaseDir() + "/views/layout/_base.html"
	templates = make(map[string]*template.Template)
	Fn := BaseDir() + "/views/home/index.html"
	templates["index"] = template.Must(template.ParseFiles(baseTemplate, Fn))
	templates["signin"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/account/signin.html"))
	templates["signup"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/account/signup.html"))
	templates["recipes"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/recipes/list.html"))
	templates["showRecipe"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/recipes/showRecipe.html"))
	templates["newRecipe"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/recipes/newRecipe.html"))
	templates["newCategory"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/categories/newCategory.html"))
	templates["listCategories"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/categories/listCategories.html"))
}

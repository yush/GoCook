package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

var templates map[string]*template.Template

func init() {
	loadTemplates()
}

func main() {

	router := httprouter.New()
	router.ServeFiles("/public/*filepath", http.Dir(BaseDir()+"public/"))
	router.ServeFiles("/images/*filepath", http.Dir(BaseDir()+"db/images/"))
	router.GET("/", IndexRoute)
	router.GET("/signin", SigninRoute)
	router.GET("/signup", SignupRoute)
	router.GET("/recipes", RecipesRoute)
	router.GET("/newrecipe", GetNewRecipeHandler)
	router.GET("/recipes/:id", GetRecipeHandler)
	router.POST("/recipes", PutRecipeHandler)
	router.GET("/import", ImportRoute)

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
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

func RecipesRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	recipes := GetAllRecipes(db)
	if err := templates["recipes"].Execute(res, recipes); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func GetRecipeHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	recipe := GetRecipe(db, p.ByName("id"))
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
	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	file, handler, err := req.FormFile("uploadfile")
	if err != nil {
		log.Fatal(err)
	}

	UploadRecipe(db, file, handler, req.FormValue("name"))
	http.Redirect(res, req, "/recipes", 301)
}

func ImportRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Fatal(err)
	}
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
}

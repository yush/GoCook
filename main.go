package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"html/template"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	_ "fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/yush/GoCook/model"
)

//var schemaDecoder = schema.NewDecoder()
//var sessionStore = sessions.NewCookieStore([]byte("your-secret-stuff-here"))

var templates map[string]*template.Template

func BaseDir() string {
	DirName := "/home/clem/go/src/github.com/yush/GoCook/views"
	Path := filepath.Dir(DirName)
	return Path
}

func init() {
	loadTemplates()
}

func main() {

	router := httprouter.New()
	router.ServeFiles("/public/*filepath", http.Dir("public/"))
	router.ServeFiles("/images/*filepath", http.Dir("db/images/"))
	router.GET("/", IndexRoute)
	router.GET("/signin", SigninRoute)
	router.GET("/signup", SignupRoute)
	router.GET("/recipes", RecipesRoute)
	router.GET("/import", ImportRoute)

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func IndexRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if err := templates["index"].Execute(res, nil); err != nil{
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func AboutRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if err := templates["about"].Execute(res, nil); err != nil{
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func ContactRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if err := templates["contact"].Execute(res, nil); err != nil{
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func SigninRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := templates["signin"].Execute(res, nil); err != nil{
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func SignupRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := templates["signup"].Execute(res, nil); err != nil{
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func RecipesRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	recipes := model.GetAllRecipes(db)
	if err := templates["recipes"].Execute(res, recipes); err != nil{
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const DIR_IMPORT = "/db/images/import/"
const DIR_ORIGINAL = "/db/images/original/"

func ImportRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	existingFiles, err :=  ioutil.ReadDir(BaseDir() + DIR_ORIGINAL)
	checkErr(err)

	files, err := ioutil.ReadDir(BaseDir() + DIR_IMPORT)
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
						
			model.Insert(db, file.Name())
			// Read all content of src to data
			data, err := ioutil.ReadFile(BaseDir() + DIR_IMPORT+ file.Name())
			checkErr(err)
			// Write data to dst
			err = ioutil.WriteFile(BaseDir() + DIR_ORIGINAL + file.Name(), data, 0644)
			checkErr(err)
		}
		os.Remove(BaseDir() + DIR_IMPORT+ file.Name())
	}
	http.Redirect(res, req, "/recipes", 301)
}

func loadTemplates(){
	var baseTemplate = BaseDir()+"/views/layout/_base.html"
	templates = make(map[string]*template.Template)
	Fn := BaseDir()+"/views/home/index.html"
	templates["index"] = template.Must(template.ParseFiles(baseTemplate, Fn,))
	templates["signin"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/account/signin.html",))
	templates["signup"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/account/signup.html",))
	templates["recipes"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/recipes/list.html",))
}
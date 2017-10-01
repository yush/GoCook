package main

import (
	"database/sql"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"html"
	"html/template"
	"image"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var templates map[string]*template.Template
var sessions *SessionManagerMem

func init() {
	loadTemplates()
	sessions = new(SessionManagerMem)
	sessions.Sessions = make(map[string]Session)
}

func main() {

	router := httprouter.New()
	router.ServeFiles("/public/*filepath", http.Dir(BaseDir()+"public/"))
	router.ServeFiles("/images/*filepath", http.Dir(BaseDir()+"db/images/"))
	router.GET("/import", ImportRoute)
	router.GET("/out", LogoutUser)
	router.GET("/signin", SigninRoute)
	router.POST("/signin", LoginUser)
	router.GET("/signup", SignupRoute)
	router.POST("/signup", NewUser)
	router.GET("/categories", ListCategories)
	router.GET("/newcategory", NewCategories)
	router.POST("/categories", PostNewCategories)
	router.GET("/recipes", RecipesRoute)
	router.POST("/deleterecipes", DeleteRecipesRoute)
	router.GET("/newrecipe", GetNewRecipeHandler)
	router.GET("/recipes/:id", GetRecipeHandler)
	router.POST("/recipes", PutRecipeHandler)
	router.GET("/", IndexRoute)

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Println("ListenAndServe: ", err.Error())
	}
}

func getDb() *sql.DB {
	db, err := sql.Open("sqlite3", BaseDir()+"/db/gocook.db3")
	if err != nil {
		log.Println(err)
	}
	return db
}

func IndexRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	data := struct {
		ASession *Session
	}{
		nil,
	}
	if err := templates["index"].Execute(res, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func SigninRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	data := struct {
		ASession *Session
		Message  string
	}{
		nil,
		"",
	}
	if err := templates["signin"].Execute(res, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func LoginUser(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	user := GetUserByEmail(getDb(), req.FormValue("email"))
	if user != nil {
		session, _ := sessions.AddNew(user)
		cookie := http.Cookie{Name: "Cookbook", Value: html.EscapeString(session.SessionID), Path: "/", HttpOnly: true, MaxAge: 300}
		http.SetCookie(res, &cookie)
		http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
	} else {
		redirectToLogin(res)
	}
}

func redirectToLogin(res http.ResponseWriter) {
	data := struct {
		ASession *Session
		Message  string
	}{
		nil,
		"User not found or wrong password",
	}
	if err := templates["signin"].Execute(res, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func SignupRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := templates["signup"].Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func LogoutUser(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	s, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		sessions.Remove(s.Email)
		cookie := http.Cookie{Name: "Cookbook", Value: "", Path: "/", HttpOnly: true, MaxAge: 300}
		http.SetCookie(res, &cookie)
	}
	http.Redirect(res, req, "/", http.StatusMovedPermanently)
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
	http.Redirect(res, req, "/categories", http.StatusMovedPermanently)
}

func RecipesRoute(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	res.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 1))
	user, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		db := getDb()
		defer db.Close()

		recipes := GetAllRecipes(db, user.UserId)
		Sess, err := sessions.LoggedInUser(req)
		if err != nil {
			redirectToLogin(res)
		} else {

			data := struct {
				Recipes  []Recipe
				ASession *Session
			}{
				recipes,
				Sess,
			}

			if err := templates["recipes"].Execute(res, data); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
		}
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
				log.Println(err)
			}
			recipe := GetRecipe(db, idRecipe)
			RemoveFile(recipe.Filepath)
			DeleteRecipe(db, idRecipe)
		}
	}
	http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
}

func GetRecipeHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	db := getDb()
	defer db.Close()

	idRecipe, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		log.Println(err)
	}

	recipe := GetRecipe(db, idRecipe)

	sess, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {

		data := struct {
			ARecipe  Recipe
			ASession *Session
		}{
			recipe,
			sess,
		}

		if err := templates["showRecipe"].Execute(res, data); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	}
}

func GetNewRecipeHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		data := struct {
			ASession *Session
		}{
			session,
		}
		if err := templates["newRecipe"].Execute(res, data); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	}
}

func PutRecipeHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	db := getDb()
	defer db.Close()

	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {

		file, handler, err := req.FormFile("uploadfile")
		if err != nil {
			log.Println(err)
		}

		img, _, errDecode := image.Decode(file)
		if errDecode != nil {
			log.Println(errDecode)
		}
		UploadRecipe(db, session.UserId, img, handler, req.FormValue("name"))
		http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
	}
}

func ImportRoute(res http.ResponseWriter, req *http.Request, par httprouter.Params) {
	db := getDb()
	defer db.Close()

	user, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		ImportRecipes(db, user.UserId, BaseDir()+DirFileStorage())
		RecipesRoute(res, req, par)
	}
}

func NewUser(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db := getDb()
	defer db.Close()

	CreateNewUser(db, req.FormValue("email"), req.FormValue("pass"), req.FormValue("passConf"))
	http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
}

func loadTemplates() {
	var baseTemplate = BaseDir() + "/views/layout/_base.html"
	templates = make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/home/index.html"))
	templates["signin"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/account/signin.html"))
	templates["signup"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/account/signup.html"))
	templates["recipes"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/recipes/list.html"))
	templates["showRecipe"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/recipes/showRecipe.html"))
	templates["newRecipe"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/recipes/newRecipe.html"))
	templates["newCategory"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/categories/newCategory.html"))
	templates["listCategories"] = template.Must(template.ParseFiles(baseTemplate, BaseDir()+"/views/categories/listCategories.html"))
}

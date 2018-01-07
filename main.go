package main

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"image"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jasonlvhit/gocron"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var templates map[string]*template.Template
var sessions *SessionManagerMem
var Sqlite3conn []*sqlite3.SQLiteConn

func init() {
	sql.Register("sqlite3_backup",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				Sqlite3conn = append(Sqlite3conn, conn)
				return nil
			},
		})

	loadTemplates()
	sessions = new(SessionManagerMem)
	sessions.Sessions = make(map[string]Session)
}

func main() {

	router := mux.NewRouter()

	// login
	router.HandleFunc("/signin", SigninRoute).Methods("GET")
	router.HandleFunc("/out", LogoutUser).Methods("GET")
	router.HandleFunc("/signin", LoginUser).Methods("POST")
	router.HandleFunc("/signup", SignupRoute).Methods("GET")
	router.HandleFunc("/signup", NewUser).Methods("POST")

	// recipes
	router.HandleFunc("/recipes", RecipesRoute).Methods("GET")
	router.HandleFunc("/newrecipe", GetNewRecipeHandler).Methods("GET")
	router.HandleFunc("/recipes/{id}", GetRecipeHandler).Methods("GET")
	router.HandleFunc("/recipes/{id}/image", GetRecipeImageHandler).Methods("GET")
	router.HandleFunc("/recipes", PutRecipeHandler).Methods("POST")
	router.HandleFunc("/deleterecipes", DeleteRecipesRoute).Methods("POST")

	// categories
	router.HandleFunc("/categories/{id}", ListRecipesByCategories).Methods("GET")
	router.HandleFunc("/categories", ListCategories).Methods("GET")
	router.HandleFunc("/newcategory", NewCategories).Methods("GET")
	router.HandleFunc("/categories", PostNewCategories).Methods("POST")
	router.HandleFunc("/backup", BackupRoute).Methods("GET")

	// services
	router.HandleFunc("/import", ImportRoute).Methods("GET")

	fsPublic := http.FileServer(http.Dir("./public"))
	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fsPublic))

	// default route
	router.HandleFunc("/", IndexRoute).Methods("GET")

	gocron.Every(1).Day().At("05:00").Do(BackupToFTP)
	gocron.Start()

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Println("ListenAndServe: ", err.Error())
	}
}

func getDb() *sql.DB {
	db, err := sql.Open("sqlite3", DatabasePath())
	if err != nil {
		log.Println(err)
	}
	return db
}

func IndexRoute(res http.ResponseWriter, req *http.Request) {
	s, _ := sessions.LoggedInUser(req)
	data := struct {
		ASession *Session
	}{
		s,
	}
	if err := templates["index"].Execute(res, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func SigninRoute(res http.ResponseWriter, req *http.Request) {
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

func LoginUser(res http.ResponseWriter, req *http.Request) {
	user := GetUserByEmail(getDb(), req.FormValue("email"))
	if user == nil {
		redirectToLogin(res)
	}
	if !user.IsSamePassword(req.FormValue("password")) {
		redirectToLogin(res)
	}
	session, _ := sessions.AddNew(user)
	cookie := http.Cookie{Name: "Cookbook", Value: html.EscapeString(session.SessionID), Path: "/", HttpOnly: true, MaxAge: 300}
	http.SetCookie(res, &cookie)
	http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
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

func SignupRoute(res http.ResponseWriter, req *http.Request) {
	if err := templates["signup"].Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func LogoutUser(res http.ResponseWriter, req *http.Request) {
	s, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		sessions.Remove(s.Email)
		cookie :=
			http.Cookie{Name: "Cookbook", Value: "", Path: "/", HttpOnly: true, MaxAge: 300}
		http.SetCookie(res, &cookie)
	}
	http.Redirect(res, req, "/", http.StatusMovedPermanently)
}

func NewCategories(res http.ResponseWriter, req *http.Request) {
	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		data := struct {
			ASession *Session
		}{
			session,
		}
		if err := templates["newCategory"].Execute(res, data); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	}
}

func ListCategories(res http.ResponseWriter, req *http.Request) {
	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		db := getDb()
		defer db.Close()

		categories := GetAllCategories(db, session.UserId)
		data := struct {
			Categories []Category
			ASession   *Session
		}{
			categories,
			session,
		}
		if err := templates["listCategories"].Execute(res, data); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	}
}

func PostNewCategories(res http.ResponseWriter, req *http.Request) {
	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		db := getDb()
		defer db.Close()

		NewCategory(db, req.FormValue("name"), session.UserId)
		http.Redirect(res, req, "/categories", http.StatusMovedPermanently)
	}
}

func RecipesRoute(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 1))
	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		db := getDb()
		defer db.Close()

		Sess, err := sessions.LoggedInUser(req)
		if err != nil {
			redirectToLogin(res)
		} else {
			recipes := GetAllRecipes(db, session.UserId)
			sort.Sort(ByName(recipes))
			categories := GetAllCategories(db, session.UserId)

			data := struct {
				ASession   *Session
				Recipes    []Recipe
				Categories []Category
			}{
				Sess,
				recipes,
				categories,
			}

			if err := templates["recipes"].Execute(res, data); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func ListRecipesByCategories(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 1))
	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		db := getDb()
		defer db.Close()

		Sess, err := sessions.LoggedInUser(req)
		if err != nil {
			redirectToLogin(res)
		} else {
			vars := mux.Vars(req)
			idCat, err := strconv.Atoi(vars["id"])
			if err != nil {
				println(err)
			}
			recipes := GetAllRecipesForCat(db, session.UserId, uint(idCat))
			sort.Sort(ByName(recipes))
			categories := GetAllCategories(db, session.UserId)

			data := struct {
				ASession   *Session
				Recipes    []Recipe
				Categories []Category
			}{
				Sess,
				recipes,
				categories,
			}

			if err := templates["recipes"].Execute(res, data); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func DeleteRecipesRoute(res http.ResponseWriter, req *http.Request) {
	var Action string

	s, _ := sessions.LoggedInUser(req)
	db := getDb()
	defer db.Close()

	req.ParseForm()
	for iAction := range req.PostForm {
		if strings.Contains(iAction, "delete") {
			Action = "delete"
			break
		} else if strings.Contains(iAction, "changeCat") {
			Action = "changeCat"
			break
		}
	}

	for i := range req.PostForm {
		if strings.Contains(i, "check-") {
			strIdRecipe := strings.TrimPrefix(i, "check-")
			idRecipe, err := strconv.Atoi(strIdRecipe)
			if err != nil {
				log.Println(err)
			}
			recipe := GetRecipe(db, idRecipe)
			if Action == "delete" {
				RemoveFile(recipe.Filepath)
				DeleteRecipe(db, s.UserId, uint(idRecipe))
			} else if Action == "changeCat" {
				idCat, errConv := strconv.Atoi(req.PostForm.Get("catId"))
				if errConv != nil {
					log.Println(errConv)
				}
				err = UpdateCategoryDetails(db, s.UserId, idRecipe, idCat)
				if err != nil {
					println(err)
				}
			}
		}
	}
	http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
}

func GetRecipeHandler(res http.ResponseWriter, req *http.Request) {
	sess, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		db := getDb()
		defer db.Close()

		vars := mux.Vars(req)
		idRecipe, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Println(err)
		}

		recipe := GetRecipe(db, idRecipe)
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

func GetRecipeImageHandler(res http.ResponseWriter, req *http.Request) {
	_, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		db := getDb()
		defer db.Close()

		vars := mux.Vars(req)
		idRecipe, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Println(err)
		}

		recipe := GetRecipe(db, idRecipe)
		http.ServeFile(res, req, BaseDir()+DirFileStorage()+recipe.Filepath)
	}
}

func GetNewRecipeHandler(res http.ResponseWriter, req *http.Request) {
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

func PutRecipeHandler(res http.ResponseWriter, req *http.Request) {
	db := getDb()
	defer db.Close()

	session, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		req.ParseForm()
		file, handler, err := req.FormFile("uploadfile")
		if err != nil {
			log.Println(err)
		}

		img, _, errDecode := image.Decode(file)
		if errDecode != nil {
			log.Println(errDecode)
		}
		UploadRecipe(db, session.UserId, img, handler.Filename, req.FormValue("name"))
		http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
	}
}

func ImportRoute(res http.ResponseWriter, req *http.Request) {
	db := getDb()
	defer db.Close()

	user, err := sessions.LoggedInUser(req)
	if err != nil {
		redirectToLogin(res)
	} else {
		err = ImportRecipes(db, user.UserId, BaseDir()+DirFileStorage())
		if err != nil {
			println(err)
		}
		RecipesRoute(res, req)
	}
}

func NewUser(res http.ResponseWriter, req *http.Request) {
	db := getDb()
	defer db.Close()

	CreateNewUser(db, req.FormValue("email"), req.FormValue("pass"), req.FormValue("passConf"))
	http.Redirect(res, req, "/recipes", http.StatusMovedPermanently)
}

func BackupRoute(res http.ResponseWriter, req *http.Request) {
	err := BackupDb("test")
	if err != nil {
		log.Println(err)
	}
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

package model

import (
	"database/sql"
	"log"
)

type Recipe struct {
	Id uint
	Name string
	Filepath string
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

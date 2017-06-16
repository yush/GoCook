package model

import (
	"database/sql"
	"log"
)

type Recipe struct {
	Id uint
	Name string
}

func GetAllRecipes(db *sql.DB) []Recipe {
	recipes := make([]Recipe, 0, 10)
	db.Begin()
	rows, err := db.Query("select name from recipes")

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.Name); err != nil {
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

func Insert(db *sql.DB, name string) {
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
	_, err = stmt.Exec(name, name)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()	
}

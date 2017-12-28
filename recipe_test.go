package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"image"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const TestDbName = "/db/gocook.test.db3"

func TestResizeRecipeLandscape(t *testing.T) {
	imgIn := image.NewRGBA(image.Rect(0, 0, 2000, 200))
	assert.Equal(t, image.Rect(0, 0, 2000, 200), imgIn.Bounds())
	imgOut := resizeFile(imgIn, 1000)
	assert.Equal(t, image.Rect(0, 0, 1000, 100), imgOut.Bounds())
}

func TestResizeRecipePortrait(t *testing.T) {
	imgIn := image.NewRGBA(image.Rect(0, 0, 200, 800))
	assert.Equal(t, image.Rect(0, 0, 200, 800), imgIn.Bounds())
	imgOut := resizeFile(imgIn, 600)
	assert.Equal(t, image.Rect(0, 0, 150, 600), imgOut.Bounds())
}

func createDatabase() {
	os.Create(BaseDir() + TestDbName)

	commands := []string{
		"go build -i -o ./tests/goose bitbucket.org/liamstask/goose/cmd/goose",
		"./tests/goose -env test up",
	}

	for _, cmd := range commands {
		args := strings.Split(cmd, " ")
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			log.Fatalf("%s:\n%v\n\n%s", err, cmd, out)
		}
	}
}

func destroyDatabase() {
	os.Remove(BaseDir() + TestDbName)
}

func TestMain(m *testing.M) {
	createDatabase()

	ret := m.Run()

	destroyDatabase()
	os.Exit(ret)
}

func getTestDb() *sql.DB {
	db, err := sql.Open("sqlite3", BaseDir()+TestDbName)
	if err != nil {
		log.Println(err)
	}
	return db
}

func TestInsert(t *testing.T) {
	db := getTestDb()
	defer db.Close()
	NewRecipeId := Insert(db, 0, "Recipe1", "test/img1.jpg")
	recipe := GetRecipe(db, int(NewRecipeId))
	assert.Equal(t, "Recipe1", recipe.Name)
}

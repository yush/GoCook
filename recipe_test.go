package main

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"image"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func init() {
	SetTestMode()
}

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
	os.Create(DatabasePath())

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
	os.Remove(DatabasePath())
}

func TestMain(m *testing.M) {
	createDatabase()

	ret := m.Run()

	destroyDatabase()
	os.Exit(ret)
}

func TestInsert(t *testing.T) {
	db := getDb()
	defer db.Close()
	NewRecipeID := Insert(db, 0, "Recipe1", "test/img1.jpg")
	recipe := GetRecipe(db, int(NewRecipeID))
	assert.Equal(t, "Recipe1", recipe.Name)
}

func TestUploadRecipe(t *testing.T) {
	db := getDb()
	defer db.Close()

	f, err := os.Open(BaseDir() + "tests/images/img1.jpg")
	if err != nil {
		t.Error(err)
	}
	img, _, errDecode := image.Decode(bufio.NewReader(f))
	if errDecode != nil {
		t.Error(err)
	}
	newID := UploadRecipe(db, 0, img, "img1.jpg", "recipe1")
	newRecipe := GetRecipe(db, newID)
	assert.Equal(t, "recipe1", newRecipe.Name)
}

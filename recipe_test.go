package main

import (
	"github.com/stretchr/testify/assert"
	"image"
	"testing"
)

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

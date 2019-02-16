package main

import (
	"testing"
)

func TestEnvironment(t *testing.T) {
	Config.Set("path", "/home")
	if Config.Get("path") != "/home" {
		t.Error("Config loading failed")
	}
}

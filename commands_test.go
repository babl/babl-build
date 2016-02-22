package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestId(t *testing.T) {
	path := filepath.Join(testModuleDir(), "string-upcase")
	if err := os.Chdir(path); err != nil {
		panic(err)
	}

	// conf()
	id := id()
	expected := "larskluge/string-upcase"

	if expected != id {
		t.Errorf("Module id mismatch: want %s; got %s", expected, id)
	}
}

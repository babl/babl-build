package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const testModuleDir = "test-modules/larskluge"

func TestConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	modules, _ := filepath.Glob(filepath.Join(wd, testModuleDir, "*"))
	for _, path := range modules {
		// cd to the module dir
		if err := os.Chdir(path); err != nil {
			panic(err)
		}

		// compare generated JSON to expected JSON
		contents, err := ioutil.ReadFile("expected-config.json")
		if err != nil {
			panic(err)
		}
		var buf bytes.Buffer
		stdout = &buf
		commands["config"].Func()
		stdout = os.Stdout
		if !bytes.Equal(contents, buf.Bytes()) {
			t.Errorf("config mismatch: want %s; got %s",
				string(contents), string(buf.Bytes()))
		}
	}
}

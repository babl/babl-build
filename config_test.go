package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const testModuleDir = "test-modules"

func TestConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	err = filepath.Walk(filepath.Join(wd, testModuleDir),
		func(path string, info os.FileInfo, err error) error {
			// only test in module dirs
			if !info.IsDir() || filepath.Base(path) == testModuleDir {
				return nil
			}
			if filepath.Base(path) == ".git" {
				return filepath.SkipDir
			}

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

			return nil
		})
	if err != nil {
		panic(err)
	}
}

package main

import (
	"path"
	"path/filepath"
	"runtime"
)

func testModuleDir() string {
	_, filename, _, _ := runtime.Caller(1)
	p, _ := filepath.Abs(path.Join(path.Dir(filename), "test-modules"))
	return p
}

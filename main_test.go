package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func testModuleDir() string {
	_, filename, _, _ := runtime.Caller(1)
	p, _ := filepath.Abs(path.Join(path.Dir(filename), "test/fixtures/modules"))
	return p
}

func testModuleDirFor(module string) string {
	return filepath.Join(testModuleDir(), module)
}

func setupFor(module string) {
	_conf = nil // clear conf() cache
	path := testModuleDirFor(module)
	if err := os.Chdir(path); err != nil {
		panic(err)
	}
}

func execConfig(module string) bytes.Buffer {
	setupFor(module)
	var buf bytes.Buffer
	stdout = &buf
	commands["config"].Func()
	stdout = os.Stdout
	return buf
}

func execConfigParsed(module string) (c config) {
	content := execConfig(module)
	blob := content.Bytes()
	err := json.Unmarshal(blob, &c)
	check(err)
	return
}

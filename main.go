//go:generate go-bindata build-config.yml

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	flag.Parse()

	if len(flag.Args()) == 0 {
		commands["help"].Func()
	} else if cmd, ok := commands[flag.Arg(0)]; ok {
		cmd.Func(flag.Args()[1:]...)
	} else {
		fmt.Fprintf(os.Stderr, "Could not find command \"%s\".\n", flag.Arg(0))
	}
}

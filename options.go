package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

var (
	dryRun       bool
	marathonHost string
)

func help(args ...string) {
	fmt.Fprintln(os.Stderr, "Commands:")
	i, maxLen, names := 0, 0, make([]string, len(commands))
	for name := range commands {
		names[i] = name
		if len(name) > maxLen {
			maxLen = len(name)
		}
		i++
	}
	sort.Sort(sort.StringSlice(names))
	formatString := fmt.Sprintf("  %%s %%-%ds  # %%s\n", maxLen)
	for _, name := range names {
		fmt.Fprintf(os.Stderr, formatString, os.Args[0], name,
			commands[name].Desc)
	}
	fmt.Fprintln(os.Stderr)

	fmt.Fprintln(os.Stderr, "Options:")
	maxLen = 0
	flag.VisitAll(func(f *flag.Flag) {
		if len(f.Name)+4 > maxLen {
			maxLen = len(f.Name) + 4
		}
	})
	formatString = fmt.Sprintf("  %%-%ds  # Default: %%s\n",
		maxLen)
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(os.Stderr, formatString, "[--"+f.Name+"]",
			f.DefValue)
	})
	fmt.Fprintln(os.Stderr)
}

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "")
	flag.StringVar(&marathonHost, "marathon-host", "127.0.0.1", "")
	flag.Usage = func() {
		help()
	}
}

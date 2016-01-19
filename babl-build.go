package main

//go:generate go-bindata -nocompress build-config.yml

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

type config struct {
	Id        string `yaml:"id" json:"id"`
	Container struct {
		Type   string `yaml:"type" json:"type"`
		Docker struct {
			Image          string `yaml:"image" json:"image"`
			ForcePullImage bool   `yaml:"forcePullImage" json:"forcePullImage"`
			Network        string `yaml:"network" json:"network"`
		} `yaml:"docker" json:"docker"`
		Options []string `yaml:"options" json:"options,omitempty"`
		Volumes []struct {
			HostPath      string `yaml:"hostPath" json:"hostPath"`
			ContainerPath string `yaml:"containerPath" json:"containerPath"`
			Mode          string `yaml:"mode" json:"mode"`
		} `yaml:"volumes" json:"volumes,omitempty"`
	} `yaml:"container" json:"container"`
	Instances int      `yaml:"instances" json:"instances"`
	Cpus      float64  `yaml:"cpus" json:"cpus"`
	Mem       float64  `yaml:"mem" json:"mem"`
	Uris      []string `yaml:"uris" json:"uris"`
	Env       struct {
		ServiceTags string `yaml:"SERVICE_TAGS" json:"SERVICE_TAGS"`
		BablModule  string `yaml:"BABL_MODULE" json:"BABL_MODULE"`
		BablCommand string `yaml:"BABL_COMMAND" json:"BABL_COMMAND"`
	} `yaml:"env" json:"env"`
	Cmd string `yaml:"cmd" json:"cmd"`
}

var (
	// options
	dryRun       bool
	marathonHost string

	// YAML "overwrites"
	overwrites config
)

// auxiliary functions

func getOutput(cmd string, args ...string) string {
	output, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(output)
}

// non-command functions

func module() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Base(dir)
}

func id() string {
	return fmt.Sprintf("%s-%s", _type(), module())
}

func _type() string {
	if tags := overwrites.Env.ServiceTags; tags != "" {
		return tags
	}
	return "babl"
}

func containerOptions() []string {
	if opts := overwrites.Container.Options; opts != nil {
		return opts
	}
	return []string{}
}

func version() string {
	return "v" + strings.TrimSpace(getOutput(
		"git", "rev-list", "HEAD", "--count"))
}

func image() string {
	return fmt.Sprintf("registry.babl.sh:5000/%s:%s", id(), version())
}

func imageLatest() string {
	return regexp.MustCompile(":[^:]+$").ReplaceAllString(image(), ":latest")
}

func conf() config {
	var c config
	contents, err := Asset("build-config.yml")
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(contents, &c); err != nil {
		panic(err)
	}
	if contents, err = ioutil.ReadFile(".babl-build.yml"); err == nil {
		var local config
		if err := yaml.Unmarshal(contents, &local); err != nil {
			log.Fatal(err)
		}
		if err := mergo.MergeWithOverwrite(&c, local); err != nil {
			panic(err)
		}
	}
	c.Id = id()
	c.Container.Docker.Image = image()
	c.Env.BablModule = module()
	return c
}

func execute(name string, args ...string) {
	fmt.Println(name + " " + strings.Join(args, " "))
	if !dryRun {
		cmd := exec.Command(name, args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.Sys().(syscall.WaitStatus).ExitStatus())
			} else {
				log.Fatal(err)
			}
		}
	}
}

type command struct {
	Desc string
	Func func(...string)
}

var commands map[string]command

func init() {
	// init logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// init overwrites
	contents, err := ioutil.ReadFile(".babl-build.yml")
	if err == nil {
		_ = yaml.Unmarshal(contents, &overwrites) // ignore error
	}

	// init commands
	commands = map[string]command{
		"build": {
			"Build Docker image",
			func(args ...string) {
				cmd := append([]string{"build", "-t", image()}, args...)
				cmd = append(cmd, ".")
				execute("docker", cmd...)
				execute("docker", "tag", "-f", image(), imageLatest())
			},
		},
		"version": {
			"Print the current version to be deployed",
			func(args ...string) {
				fmt.Println(version())
			},
		},
		"config": {
			"Print the Marathon JSON config",
			func(args ...string) {
				err := json.NewEncoder(os.Stdout).Encode(conf())
				if err != nil {
					panic(err)
				}
			},
		},
		"push": {
			"Push Docker image to remote registry",
			func(args ...string) {
				execute("docker", "push", image())
				execute("docker", "push", imageLatest())
			},
		},
		"deploy": {
			"Deploy a Babl module",
			func(args ...string) {
				body := bytes.NewBuffer([]byte{})
				if err := json.NewEncoder(body).Encode(conf()); err != nil {
					panic(err)
				}
				req, err := http.NewRequest("POST",
					fmt.Sprintf("http://%s:8080/v2/apps", marathonHost), body)
				if err != nil {
					log.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := (&http.Client{}).Do(req)
				if err != nil {
					log.Fatal(err)
				}

				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					_, _ = io.Copy(os.Stdout, resp.Body) // ignore error
				}
				_ = resp.Body.Close() // ignore error
			},
		},
		"destroy": {
			"Destroy a Babl module",
			func(args ...string) {
				req, err := http.NewRequest("DELETE",
					fmt.Sprintf("http://%s:8080/v2/apps/%s",
						marathonHost, id()), nil)
				if err != nil {
					log.Fatal(err)
				}
				resp, err := (&http.Client{}).Do(req)
				if err != nil {
					log.Fatal(err)
				}
				_, _ = io.Copy(os.Stdout, resp.Body) // ignore error
				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					log.Fatalf("HTTP DELETE request returned %s", resp.Status)
				}
			},
		},
		"dist": {
			"build & push & destroy & deploy",
			func(args ...string) {
				commands["build"].Func()
				commands["push"].Func()
				commands["destroy"].Func()
				commands["deploy"].Func()
			},
		},
		"play": {
			"Play (run) a local built Babl module",
			func(args ...string) {
				execArgs := []string{"docker", "run", "-it", "--rm", "-p",
					"4444:4444", "-e", "BABL_MODULE=" + module(), "-e",
					"BABL_COMMAND=/bin/app"}
				execArgs = append(execArgs, containerOptions()...)
				execArgs = append(execArgs, image())
				execute(execArgs[0], execArgs[1:]...)
			},
		},
		"sh": {
			"Run the container with a shell",
			func(args ...string) {
				execArgs := []string{"docker", "run", "-it", "--rm", "-p",
					"4444:4444", "-e", "BABL_MODULE=" + module(), "-e",
					"BABL_COMMAND=/bin/app"}
				execArgs = append(execArgs, containerOptions()...)
				execArgs = append(execArgs, image(), "sh")
				execute(execArgs[0], execArgs[1:]...)
			},
		},
		"help": {
			"Describe available commands",
			func(args ...string) {
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
			},
		},
	}

	// init options
	flag.BoolVar(&dryRun, "dry-run", false, "")
	flag.StringVar(&marathonHost, "marathon-host", "127.0.0.1", "")
	flag.Usage = func() {
		commands["help"].Func()
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		commands["help"].Func()
	} else if cmd, ok := commands[flag.Arg(0)]; ok {
		cmd.Func(flag.Args()[1:]...)
	} else {
		fmt.Fprintf(os.Stderr, "Could not find command \"%s\".\n", flag.Arg(0))
	}
}

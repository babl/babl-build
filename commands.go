package main

//go:generate go-bindata -nocompress build-config.yml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

const retries = 3

var stdout io.Writer = os.Stdout // allow reassignment
var _conf *config = nil          // cache conf()'s result

// auxiliary functions

func conf() config {
	if _conf != nil {
		return *_conf
	}

	var c config
	contents, err := Asset("build-config.yml")
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(contents, &c); err != nil {
		panic(err)
	}
	if contents, err = ioutil.ReadFile("babl.yml"); err != nil {
		log.Fatal(err)
	}

	var local config
	if err := yaml.Unmarshal(contents, &local); err != nil {
		log.Fatal(err)
	}
	if err := mergo.MergeWithOverwrite(&c, local); err != nil {
		panic(err)
	}

	// Ugly hack to support unlimited memory usage / zero value
	if local.Mem != nil && *local.Mem == 0 {
		*c.Mem = 0
	}

	if c.Env.ServiceTags == "web" {
		reg := regexp.MustCompile(":4[0-9]+$")
		for i := 0; i < len(c.Container.Docker.Parameters); i++ {
			param := &c.Container.Docker.Parameters[i]
			if param.Key == "log-opt" {
				param.Value = reg.ReplaceAllString(param.Value, ":4990")
			}
		}
	}

	_conf = &c
	c.Container.Docker.Image = image()
	c.Env.BablModule = module()
	c.Env.BablModuleVersion = version()
	return c
}

func containerOptions() []string {
	if opts := overwrites.Container.Options; opts != nil {
		return opts
	}
	return []string{}
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

func getOutput(cmd string, args ...string) string {
	output, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(output)
}

func id() string {
	return conf().Id
}

func image() string {
	return fmt.Sprintf("registry.babl.sh:5000/%s:%s", id(), version())
}

func imageLatest() string {
	return regexp.MustCompile(":[^:]+$").ReplaceAllString(image(), ":latest")
}

func module() string {
	return id()
}

func _type() string {
	if tags := overwrites.Env.ServiceTags; tags != "" {
		return tags
	}
	return "babl"
}

func version() string {
	output, err := exec.Command("git", "rev-list", "HEAD", "--count").Output()
	if err == nil {
		return "v" + strings.TrimSpace(string(output))
	} else {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Sys().(syscall.WaitStatus).ExitStatus() == 128 {
				return "v0"
			}
		}
	}
	log.Fatal(err)
	return "fail"
}

// commands proper

type command struct {
	Desc string
	Func func(...string)
}

var commands map[string]command

func init() {
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
				err := json.NewEncoder(stdout).Encode(conf())
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
				for i := 0; i < retries; i++ {
					body := bytes.NewBuffer([]byte{})
					err := json.NewEncoder(body).Encode(conf())
					if err != nil {
						panic(err)
					}
					req, err := http.NewRequest("POST",
						fmt.Sprintf("http://%s:8080/v2/apps", marathonHost),
						body)
					if err != nil {
						log.Fatal(err)
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := (&http.Client{}).Do(req)
					if err != nil {
						log.Fatal(err)
					}

					_, _ = io.Copy(stdout, resp.Body) // ignore error
					_ = resp.Body.Close()             // ignore error
					if resp.StatusCode < 200 || resp.StatusCode >= 300 {
						log.Printf("HTTP POST request returned %s",
							resp.Status)
					}
					if resp.StatusCode == http.StatusConflict && i < retries {
						log.Print("Retrying in 1 second...")
						time.Sleep(time.Second)
						continue
					}
					break
				}
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
				_, _ = io.Copy(stdout, resp.Body) // ignore error
				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					log.Fatalf("HTTP DELETE request returned %s", resp.Status)
				}
			},
		},
		"restart": {
			"restart all instances of this module",
			func(args ...string) {
				commands["destroy"].Func()
				commands["deploy"].Func()
			},
		},
		"play": {
			"Play (run) a local built Babl module",
			func(args ...string) {
				execArgs := []string{"docker", "run", "-it", "--rm", "-p",
					"4444:4444",
					"-e", "PORT=4444",
					"-e", "BABL_MODULE=" + module(),
					"-e", "BABL_MODULE_VERSION=" + version(),
					"-e", "BABL_COMMAND=/bin/app",
				}
				execArgs = append(execArgs, containerOptions()...)
				execArgs = append(execArgs, image())
				execute(execArgs[0], execArgs[1:]...)
			},
		},
		"sh": {
			"Run the container with a shell",
			func(args ...string) {
				execArgs := []string{"docker", "run", "-it", "--rm", "-p", "4444:4444",
					"-e", "BABL_MODULE=" + module(),
					"-e", "BABL_MODULE_VERSION=" + version(),
					"-e", "BABL_COMMAND=/bin/app",
				}
				execArgs = append(execArgs, containerOptions()...)
				execArgs = append(execArgs, image(), "sh")
				execute(execArgs[0], execArgs[1:]...)
			},
		},
		"help": {
			"Describe available commands",
			help,
		},
	}
}

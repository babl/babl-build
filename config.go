package main

import (
	"io/ioutil"

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
			PortMappings   []struct {
				ContainerPort int    `yaml:"containerPort" json:"containerPort,omitempty"`
				HostPort      int    `yaml:"hostPort" json:"hostPort"`
				ServicePort   int    `yaml:"servicePort" json:"servicePort,omitempty"`
				Protocol      string `yaml:"protocol" json:"protocol,omitempty"`
			} `yaml:"portMappings" json:"portMappings,omitempty"`
			Parameters []struct {
				Key   string `yaml:"key" json:"key"`
				Value string `yaml:"value" json:"value"`
			} `yaml:"parameters" json:"parameters,omitempty"`
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
	Mem       *float64 `yaml:"mem" json:"mem"`
	Uris      []string `yaml:"uris" json:"uris"`
	Env       struct {
		ServiceTags       string `yaml:"SERVICE_TAGS" json:"SERVICE_TAGS"`
		BablModule        string `yaml:"BABL_MODULE" json:"BABL_MODULE"`
		BablModuleVersion string `yaml:"BABL_MODULE_VERSION" json:"BABL_MODULE_VERSION"`
		BablCommand       string `yaml:"BABL_COMMAND" json:"BABL_COMMAND"`
	} `yaml:"env" json:"env"`
	Cmd string `yaml:"cmd" json:"cmd"`
}

var overwrites config

func init() {
	contents, err := ioutil.ReadFile("babl.yml")
	if err == nil {
		_ = yaml.Unmarshal(contents, &overwrites) // ignore error
	}
}

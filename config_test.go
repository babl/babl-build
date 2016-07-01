package main

import (
	"testing"
)

func TestModuleId(t *testing.T) {
	c := execConfigParsed("string-upcase")
	expected := "larskluge/string-upcase"
	actual := c.Id
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestCustomizedMem(t *testing.T) {
	c := execConfigParsed("image-resize")
	expected := 128.0
	actual := *c.Mem
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestDefaultMem(t *testing.T) {
	c := execConfigParsed("string-upcase")
	expected := 16.0
	actual := *c.Mem
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestUnlimitedMem(t *testing.T) {
	c := execConfigParsed("unlimited-mem")
	expected := 0.0
	actual := *c.Mem
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestCustomizedCpus(t *testing.T) {
	c := execConfigParsed("image-resize")
	expected := 0.6
	actual := c.Cpus
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestKafkaBrokers(t *testing.T) {
	c := execConfigParsed("string-upcase")
	expected := "queue.babl.sh:9092"
	actual := c.Env.BablKafkaBrokers
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestDefaultDockerForcePullDisabled(t *testing.T) {
	c := execConfigParsed("string-upcase")
	expected := false
	actual := c.Container.Docker.ForcePullImage
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestDefaultNetwork(t *testing.T) {
	c := execConfigParsed("string-upcase")
	expected := "BRIDGE"
	actual := c.Container.Docker.Network
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestDockerImage(t *testing.T) {
	c := execConfigParsed("string-upcase")
	expected := "registry.babl.sh/larskluge/string-upcase:v20"
	actual := c.Container.Docker.Image
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

func TestVolumeMounts(t *testing.T) {
	c := execConfigParsed("babl-build")
	expected := "/usr/lib64/libsystemd.so.0"
	actual := c.Container.Volumes[4].HostPath
	if expected != actual {
		t.Errorf("config mismatch: want %s; got %s", expected, actual)
	}
}

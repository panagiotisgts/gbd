package gbd

import (
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/yaml.v3"
)

const (
	networkReplaceId string = "{NETWORK_ID}"
)

type Dependency struct {
	Image         string              `yaml:"image"`
	Version       string              `yaml:"version"`
	Name          string              `yaml:"name,omitempty"`
	ReplaceConfig []ConfigReplacement `yaml:"replaceConfig,omitempty"`
	// TODO Parse Replacement and add support for ContainerDerivedValue
	Env         map[string]string `yaml:"env,omitempty"`
	Files       []File            `yaml:"files,omitempty"`
	ExposePorts []string          `yaml:"exposePorts,omitempty"`
	Alias       string            `yaml:"alias,omitempty"`
	Build       *DockerBuild      `yaml:"build,omitempty"`
	WaitFor     WaitFor           `yaml:"waitFor,omitempty"`
}

type DockerBuild struct {
	Dockerfile string             `yaml:"dockerfile"`
	BuildArgs  map[string]*string `yaml:"buildArgs"`
	BuildLog   bool               `yaml:"buildLog"`
}

// ConfigReplacement is a struct that represents a key/value pair that will be replaced in the config file
// Key can be a dot separated path to the value if it is nested
type ConfigReplacement struct {
	ConfigOriginPath string        `yaml:"config_origin_path,omitempty"`
	TargetPath       string        `yaml:"target_path,omitempty"`
	Replacements     []Replacement `yaml:"replacements,omitempty"`
	hostFile         string        `yaml:"host_file,omitempty"`
}

type Replacement struct {
	Key   string `yaml:"key"`
	Value any    `yaml:"value"`
}

type File struct {
	TargetPath   string `yaml:"targetPath"`
	Mode         int64  `yaml:"mode"`
	Content      []byte `yaml:"content,omitempty"`
	HostFilePath string `yaml:"hostFilePath,omitempty"`
}

type ContainerDerivedValue struct {
	FromContainer         string `yaml:"fromContainer"`
	ContainerPropertyPath string `yaml:"propertyName"`
}

// WaitFor is a struct that represents a wait strategy for a container.
// Strategies Supported http, port, log, healthcheck
type WaitFor struct {
	Strategy        string        `yaml:"strategy"`
	WaitForStrategy wait.Strategy `yaml:"waitForStrategy"`
}

func (w *WaitFor) UnmarshalYAML(value *yaml.Node) error {
	vmap := make(map[string]any)
	if err := value.Decode(&vmap); err != nil {
		return err
	}
	b, err := yaml.Marshal(vmap["waitForStrategy"])
	if err != nil {
		return err
	}
	switch vmap["strategy"].(string) {
	case "log":
		var str wait.LogStrategy
		if err := yaml.Unmarshal(b, &str); err != nil {
			return err
		}
		w.Strategy = "log"
		w.WaitForStrategy = &str
	case "http":
		w.Strategy = "http"
		var str wait.HTTPStrategy
		if err := yaml.Unmarshal(b, &str); err != nil {
			return err
		}
		w.WaitForStrategy = &str
	case "healthcheck":
		w.Strategy = "healthcheck"
		w.WaitForStrategy = wait.ForHealthCheck()
	case "port":
		w.Strategy = "port"
		var str wait.HostPortStrategy
		if err := yaml.Unmarshal(b, &str); err != nil {
			return err
		}
		w.WaitForStrategy = &str

	}
	return nil
}

type StackComponent struct {
	container      testcontainers.Container `yaml:"-"`
	ContainerId    string                   `yaml:"containerId"`
	Name           string                   `yaml:"name"`
	Image          string                   `yaml:"image"`
	Version        string                   `yaml:"version"`
	Networks       []string                 `yaml:"networks"`
	NetworkAliases map[string][]string      `yaml:"networkAliases"`
	InternalIP     string                   `yaml:"internalIP"`
	Ports          map[string][]PortRef     `yaml:"ports"`
	MappedPorts    map[string]string        `yaml:"mappedPorts"`
}

type PortRef struct {
	HostIp string `yaml:"hostIp"`
	Port   string `yaml:"port"`
}

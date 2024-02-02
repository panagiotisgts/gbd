package gbd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"gopkg.in/yaml.v3"

	"github.com/PanagiotisGts/gbd/internal/utils"
)

type Env struct {
	ContextDir   string       `yaml:"context"`
	Dependencies []Dependency `yaml:"dependencies"`
	Network      string       `yaml:"-"`
}

func newEnv(contextDir string, containers []Dependency) *Env {
	return &Env{
		Dependencies: containers,
		ContextDir:   contextDir,
	}
}

func (e *Env) Build(ctx context.Context, dumpConfig bool) (*Stack, error) {
	nw, err := network.New(ctx)
	if err != nil {
		return nil, err
	}
	stack := &Stack{
		network: nw,
	}
	stack.workDir = e.ContextDir
	stack.tempDir = filepath.Join(e.ContextDir, "gbd_temp")
	err = os.Mkdir(stack.tempDir, 0755)
	if err != nil {
		return nil, err
	}

	if dumpConfig {
		b, _ := yaml.Marshal(e)
		if err := os.WriteFile(filepath.Join(stack.workDir, "gbd_config.yaml"), b, 0644); err != nil {
			fmt.Println(err)
		}
	}

	defer os.RemoveAll(stack.tempDir)
	for i := range e.Dependencies {
		ctr := baseContainerRequest(e.Dependencies[i].Image, e.Dependencies[i].Version, e.Dependencies[i].Env)
		if e.Dependencies[i].Name != "" {
			ctr.Name = e.Dependencies[i].Name
		}

		if e.Dependencies[i].Build != nil {
			ctr.FromDockerfile = testcontainers.FromDockerfile{
				Context:       e.ContextDir,
				Dockerfile:    e.Dependencies[i].Build.Dockerfile,
				Repo:          e.Dependencies[i].Image,
				Tag:           e.Dependencies[i].Version,
				BuildArgs:     e.Dependencies[i].Build.BuildArgs,
				PrintBuildLog: e.Dependencies[i].Build.BuildLog,
				KeepImage:     false,
			}
			ctr.Image = ""
		}

		ctr.WaitingFor = e.Dependencies[i].WaitFor.WaitForStrategy
		ctr.Networks = []string{nw.Name}
		ctr.NetworkAliases = map[string][]string{nw.Name: {e.Dependencies[i].Alias}}
		ctr.Files = make([]testcontainers.ContainerFile, 0)
		ctr.ExposedPorts = e.Dependencies[i].ExposePorts

		if err := stack.replaceConfigs(e.Dependencies[i].ReplaceConfig); err != nil {
			return nil, err
		}
		for _, r := range e.Dependencies[i].ReplaceConfig {
			ctr.Files = append(ctr.Files, testcontainers.ContainerFile{
				HostFilePath:      r.hostFile,
				ContainerFilePath: r.TargetPath,
				FileMode:          0644,
			})
		}

		for _, file := range e.Dependencies[i].Files {
			if file.HostFilePath != "" {
				ctr.Files = append(ctr.Files, testcontainers.ContainerFile{
					HostFilePath:      file.HostFilePath,
					ContainerFilePath: file.TargetPath,
					FileMode:          file.Mode,
				})
			} else {
				fn := utils.ExtractFileName(file.TargetPath)
				f, err := os.Create(stack.tempDir + fn)
				if err != nil {
					panic(err)
				}
				_, err = f.Write(file.Content)
				if err != nil {
					panic(err)
				}
				ctr.Files = append(ctr.Files, testcontainers.ContainerFile{
					HostFilePath:      f.Name(),
					ContainerFilePath: file.TargetPath,
					FileMode:          file.Mode,
				})
			}

		}

		tc, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: *ctr,
			Started:          true,
		})
		if err != nil {
			return nil, err
		}
		cmp := createComponent(ctx, err, tc, e.Dependencies[i])
		stack.addComponent(cmp)
	}

	return stack, nil
}

func baseContainerRequest(image, version string, env map[string]string) *testcontainers.ContainerRequest {
	return &testcontainers.ContainerRequest{
		Image: fmt.Sprintf("%s:%s", image, version),
		Env:   env,
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.AutoRemove = false
		},
	}
}

func parseConfig(path string) (map[string]any, error) {
	var cfg map[string]any
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return nil, err
		}
	case ".json":
		if err := json.Unmarshal(b, &cfg); err != nil {
			return nil, err
		}
	case ".toml":
		/*if err := toml.Unmarshal(b, &cfg); err != nil {
			return nil, err
		}*/
	}
	return cfg, nil
}

func createComponent(ctx context.Context, err error, tc testcontainers.Container, dep Dependency) StackComponent {
	name, err := tc.Name(ctx)
	networks, err := tc.Networks(ctx)
	nwa, err := tc.NetworkAliases(ctx)
	ip, err := tc.ContainerIP(ctx)
	ports, err := tc.Ports(ctx)
	pmap := make(map[string][]PortRef)
	for p, b := range ports {
		pmap[p.Port()] = make([]PortRef, len(b))
		for i, binding := range b {
			pmap[p.Port()][i] = PortRef{HostIp: binding.HostIP, Port: binding.HostPort}
		}
	}

	mappedPorts := make(map[string]string)
	for _, port := range dep.ExposePorts {
		mappedPort, err := tc.MappedPort(ctx, nat.Port(port))
		if err != nil {
			mappedPorts[port] = ""
			continue
		}
		mappedPorts[port] = mappedPort.Port()
	}
	return StackComponent{
		container:      tc,
		Image:          dep.Image,
		Version:        dep.Version,
		ContainerId:    tc.GetContainerID(),
		Name:           strings.TrimPrefix(name, "/"),
		Networks:       networks,
		NetworkAliases: nwa,
		InternalIP:     ip,
		Ports:          pmap,
		MappedPorts:    mappedPorts,
	}
}

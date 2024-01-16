package gbd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"gopkg.in/yaml.v3"

	"github.com/PanagiotisGts/gbd/internal/utils"
)

type Stack struct {
	components []StackComponent
	network    *testcontainers.DockerNetwork
	tempDir    string
	workDir    string
}

func (s *Stack) addComponent(c StackComponent) {
	s.components = append(s.components, c)
}

func (s *Stack) Teardown(ctx context.Context) error {
	ids := make([]string, len(s.components))
	for i := range s.components {
		ids[i] = s.components[i].ContainerId
		d := 5 * time.Second
		err := s.components[i].container.Stop(ctx, &d)
		if err != nil {
			return err
		}
	}
	utils.WaitForContainerToBeRemoved(ids...)
	return s.network.Remove(ctx)
}

func (s *Stack) GetComponent(name string) (StackComponent, error) {
	for _, c := range s.components {
		if c.Name == name {
			return c, nil
		}
	}
	return StackComponent{}, fmt.Errorf("component not found")

}

func (s *Stack) Print() []byte {
	b, err := yaml.Marshal(s)
	if err != nil {
		log.Println(err)
	}
	return b
}

func (s *Stack) replaceConfigs(replacements []ConfigReplacement) error {
	for i, r := range replacements {
		cfg, err := parseConfig(s.workDir + r.ConfigOriginPath)
		if err != nil {
			return err
		}
		for _, rep := range r.Replacements {
			if dv, ok := rep.Value.(*ContainerDerivedValue); ok {
				s.replaceConfigDerivedValue(rep.Key, dv, cfg)
			} else if dv, ok := rep.Value.(map[string]any); !ok {
				s.replaceConfigString(rep.Key, rep.Value, cfg)
			} else {
				s.replaceConfigDerivedValue(rep.Key, &ContainerDerivedValue{
					FromContainer:         dv["fromContainer"].(string),
					ContainerPropertyPath: dv["propertyName"].(string),
				}, cfg)
			}
		}
		fn := filepath.Join(s.tempDir, utils.ExtractFileName(r.ConfigOriginPath))
		if err := s.flushConfig(fn, cfg); err != nil {
			return err
		}
		replacements[i].hostFile = fn
	}
	return nil
}

func (s *Stack) replaceConfigDerivedValue(key string, value *ContainerDerivedValue, cfg map[string]any) {
	var cvalue any
	for _, c := range s.components {
		if c.Name == value.FromContainer {

			path := value.ContainerPropertyPath
			if strings.Contains(path, networkReplaceId) {
				path = strings.Replace(path, networkReplaceId, fmt.Sprintf("\"%s\"", c.Networks[0]), 1)
			}

			containerCfg, err := utils.InspectContainer(context.Background(), c.Image, c.Version)
			if err != nil {
				panic(err)
			}

			cvalue, err = utils.FindValueInJson(containerCfg, path)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Replacing config '%s' from container '%s' with value '%v' mapped from '%s'\n", key, c.Name, cvalue, path)
		}
	}
	keyPath := strings.Split(key, ".")
	utils.FindAndReplace(keyPath, cvalue, cfg)
}

func (s *Stack) replaceConfigString(key string, value any, cfgMap map[string]any) {
	keyPath := strings.Split(key, ".")
	utils.FindAndReplace(keyPath, value, cfgMap)
}

func (s *Stack) flushConfig(target string, cfg map[string]any) error {
	var bytes []byte
	var err error
	switch filepath.Ext(target) {
	case ".yaml", ".yml":
		bytes, err = yaml.Marshal(cfg)
	case ".json":
		bytes, err = json.Marshal(cfg)
	case ".toml":
		//bytes, err = toml.Marshal(cfg)
	}
	if err != nil {
		return err
	}
	if err := os.WriteFile(target, bytes, 0644); err != nil {
		return err
	}
	return nil
}

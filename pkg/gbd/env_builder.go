package gbd

import (
	"os"

	"gopkg.in/yaml.v3"
)

func NewEnv(contextDir string, dependencies []Dependency) *Env {
	return newEnv(contextDir, dependencies)
}

func NewEnvFromConfig(configPath string) (*Env, error) {
	var env Env
	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

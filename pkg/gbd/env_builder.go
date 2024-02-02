package gbd

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func New(contextDir string, dependencies []Dependency) *Env {
	err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	if err != nil {
		log.Println(err)
	}
	//defer os.Unsetenv("TESTCONTAINERS_RYUK_DISABLED")
	return newEnv(contextDir, dependencies)
}

func NewFromConfig(configPath string) (*Env, error) {
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

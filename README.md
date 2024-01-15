# GBD - GoBrewDock 
###### Work In Progress

A wrapper tool of TestContainers (go) that handles building and running multiple containers 
as a declarative dependency stack. It can be used as a CLI tool or embedded as a library. The inspiration for this tool
stemmed from the concept of a 'System Under Test' (SUT), often a single (μ)service, used during testing.
There was a need for a convenient method to create and manage the SUT's dependencies, as well as its image and configuration parameters.
In that way, the SUT can be tested in isolation, without the need to manually create and manage its dependencies and as close as possible to production.

## Features
Most of the features provided by TestContainers are supported.

A unique feature provided is the ability to declare replacements for configuration parameters (files, env vars) from
previously declared containers in the stack by utilizing a JSONPath pointer to the Docker Inspect JSON.

A library generated configuration can be dumped to a `yaml` file that can be reused, modified and executed either via the CLI
tool or the library.

Another unique feature of the CLI tool is the ability to perform hot-reload of the specified config file. This is done 
either manually or when modifying the source file, resulting in the termination and the redeployment of the containers which
were part of the stack (or new ones).

## Docker Inspect JSON Path dynamic params
 - `{NETWORK_ID}` - The ID of the network the stack is deployed to (generated)

## CLI Usage

- Dry-Run :
  - Run the deployment stack for verification, followed by a cleanup afterward.  
  - gbd dry-run --config _{config.yaml}_ --context _{context_dir}_


- Watcher :
    - Run the deployment stack and watch for changes in the source file. If a change is detected, the stack is redeployed.
    - gbd watcher --config _{config.yaml}_ --context _{context_dir}_ _[--dump true | false]_


<details>
  <summary>Example config file (Same as next Go example)</summary>

```yaml
context: "/usr/projects/my_service"
dependencies:
    - image: postgres
      version: latest
      name: test-postgres
      env:
        POSTGRES_USER: admin
        POSTGRES_PASSWORD: root
        POSTGRES_DB: test_db
      exposePorts:
        - "5432"
      alias: pgtc
      waitFor:
        strategy: log
        waitForStrategy:
            log: database system is ready to accept connections
            isregexp: false
            occurrence: 1
            pollinterval: 100ms
    - image: my_service
      version: latest
      name: my-service
      replaceConfig:
        - config_origin_path: /configs/config.yaml
          target_path: /configs/config.yaml
          replacements:
            - key: db.host
              value:
                fromContainer: test-postgres
                propertyName: NetworkSettings.Networks[{NETWORK_ID}].Aliases[0]
            - key: server.address
              value: :8080
      exposePorts:
        - "8080"
      alias: my-awesome-service
      build:
        dockerfile: Dockerfile
        buildArgs:
            VERSION: latest
        buildLog: true
      waitFor:
        strategy: log
        waitForStrategy:
            log: service is ready to serve requests
            isregexp: false
            occurrence: 1
            pollinterval: 100ms
```
</details>

## Library Usage
<details>
  <summary>Sample usage with a Postgres container and a custom service</summary>

```go
package test

import (
	"context"

	"github.com/panosg/gbd/pkg/gbd"
	"github.com/testcontainers/testcontainers-go/wait"
)

func main() {
	ctx := context.Background()
	ctDir := "/usr/projects/my_service"
	e := gbd.NewEnv(ctDir, []gbd.Dependency{
		{
			Image:   "postgres",
			Version: "latest",
			Name:    "test-postgres",
			Env: map[string]string{
				"POSTGRES_USER":     "admin",
				"POSTGRES_PASSWORD": "root",
				"POSTGRES_DB":       "test_db",
			},
			ExposePorts: []string{"5432"},
			Alias:       "pgtc",
			WaitFor: gbd.WaitFor{
				Strategy:        "log",
				WaitForStrategy: wait.ForLog("database system is ready to accept connections"),
			},
		},
		{
            Image:   "my_service",
            Version: "latest",
            Name:    "my-service",
            Build: &gbd.DockerBuild{
                Dockerfile: "Dockerfile",
                BuildArgs:  map[string]*string{},
                BuildLog:   true,
            },
            ExposePorts: []string{"8080"},
            ReplaceConfig: []gbd.ConfigReplacement{
                {
                    ConfigOriginPath: "/configs/config.yaml",
                    TargetPath:       "/configs/config.yaml",
                    Replacements: []gbd.Replacement{
                        {
                            Key: "database.host",
                            Value: &gbd.ContainerDerivedValue{
                                FromContainer:         "test-postgres",
                                ContainerPropertyPath: "NetworkSettings.Networks[{NETWORK_ID}].Aliases[0]",
                            },
                        },
                        {
                            Key:   "server.address",
                            Value: ":8080",
                        },
                    },
              },
		    },
            Alias:   "my-awesome-service
            WaitFor: gbd.WaitFor{Strategy: "log", WaitForStrategy: wait.ForLog("service is ready to serve requests")},
	    },
	})

	stack, err := e.Build(ctx, true)
	if err != nil {
		panic(err)
	}

	myServiceContainer, err := stack.GetComponent("my-service")
	port := myServiceContainer.MappedPorts["8080"]
}
```
</details>
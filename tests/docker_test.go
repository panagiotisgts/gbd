

package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/PanagiotisGts/gbd/internal/utils"
	"github.com/PanagiotisGts/gbd/pkg/gbd"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

func TestContainerCreation(t *testing.T) {
	dir, err := os.Getwd()
	fmt.Println(dir)
	require.Nil(t, err)
	dir = filepath.Join(dir, "testdata")
	fmt.Println(dir)
	ctx := context.Background()
	env := gbd.New(dir, []gbd.Dependency{
		{
			Image:   "alpine",
			Version: "test",
			Name:    "alpine_spin",
			Env: map[string]string{
				"TEST_ENV_VAR": "admin",
			},
			Build: &gbd.DockerBuild{
				Dockerfile: "Dockerfile",
				BuildLog:   false,
			},
			Files: []gbd.File{
				{
					TargetPath:   "/tmp/test.txt",
					Mode:         0644,
					HostFilePath: filepath.Join(dir, "test.txt"),
				},
			},
			Alias: "test_alpine",
		},
	})

	stack, err := env.Build(ctx, false)
	require.Nil(t, err)

	t.Cleanup(func() {
		err := stack.Teardown(ctx)
		require.Nil(t, err)
	})

	_, err = stack.GetComponent("alpine_spin")
	require.Nil(t, err)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.Nil(t, err)

	id, err := utils.FindContainer("alpine", "test", cli)
	require.Nil(t, err)

	verifyFileIsCopied(t, id)
	assertEnvVariableExists(t, id)
	assertContainerProperties(t, ctx, id, cli)

}

func assertContainerProperties(t *testing.T, ctx context.Context, containerId string, cli *client.Client) {
	ins, err := cli.ContainerInspect(ctx, containerId)
	require.Nil(t, err)
	require.Contains(t, ins.Config.Env, "TEST_ENV_VAR=admin")
	require.Equal(t, "alpine_spin", ins.Name[1:])
	require.Equal(t, "alpine:test", ins.Config.Image)
	for _, v := range ins.NetworkSettings.Networks {
		require.Contains(t, v.Aliases, "test_alpine")
	}
}

func verifyFileIsCopied(t *testing.T, containerId string) {
	cmd := exec.Command("docker", "exec", containerId, "cat", "/tmp/test.txt")
	out, err := cmd.CombinedOutput()
	require.Nil(t, err)
	require.Equal(t, "Test file", string(out))
}

func assertEnvVariableExists(t *testing.T, containerId string) {
	cmd := exec.Command("docker", "exec", containerId, "sh", "-c", `echo $TEST_ENV_VAR`)
	out, err := cmd.CombinedOutput()
	require.Nil(t, err)
	require.Contains(t, string(out), "admin")
}

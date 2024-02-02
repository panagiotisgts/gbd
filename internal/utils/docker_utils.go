package utils

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func InspectContainer(ctx context.Context, image, version string) ([]byte, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	ctr, err := FindContainer(image, version, cli)
	if err != nil {
		return nil, err
	}
	_, json, err := cli.ContainerInspectWithRaw(ctx, ctr, true)
	if err != nil {
		return nil, err
	}

	return json, nil
}

func FindContainer(image, version string, cli *client.Client) (string, error) {
	f2 := filters.NewArgs(filters.KeyValuePair{Key: "ancestor", Value: fmt.Sprintf("%s:%s", image, version)})
	r, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: f2})
	if err != nil {
		return "", err
	}
	return r[0].ID, nil
}

func WaitForContainerToBeRemoved(ids ...string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	rm := make(map[string]bool)
	for _, id := range ids {
		rm[id] = false
	}
	exitFlag := false
	for {
		if exitFlag {
			return
		}
		for _, id := range ids {
			if !rm[id] {
				f := filters.NewArgs(filters.KeyValuePair{Key: "id", Value: id})
				r, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: f})
				if err != nil && r == nil {
					rm[id] = true
				}
				cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{Force: true})
			}
		}
		for _, v := range rm {
			if !v {
				exitFlag = false
				break
			}
		}
		exitFlag = true
	}
}

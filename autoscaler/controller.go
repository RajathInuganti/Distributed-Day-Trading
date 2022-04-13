package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func createContainers(ctx context.Context, cli *client.Client, envs Envs) []string {

	containers := make([]string, envs.maxWorkers)
	for i := 0; i < envs.maxWorkers; i++ {
		// Create the maximum number of containers allowed. Start and stop them as necessary
		// Yet to be implemented as the transaction worker needs to be implemented as well
	}
	return containers
}

func startContainer(ctx context.Context, cli *client.Client, ID string, envs Envs, updates chan string) {
	if err := cli.ContainerStart(ctx, ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	go monitor(ctx, cli, ID, envs, updates)
}

func stopContainer(ctx context.Context, cli *client.Client, ID string) {
	if err := cli.ContainerStop(ctx, ID, nil); err != nil {
		panic(err)
	}
}

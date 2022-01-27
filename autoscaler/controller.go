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

func updateWorkerRecords(ctx context.Context, cli *client.Client, envs Envs, container string, start bool, updates chan ContainerDetail) {

	var ID string

	if start {
		if len(RunningWorkerRecord) < envs.maxWorkers {
			ID, StoppedWorkerRecord = StoppedWorkerRecord[0], StoppedWorkerRecord[1:]
			RunningWorkerRecord[ID] = true
			startContainer(ctx, cli, ID, envs, updates)
			return
		}
	}

	if len(StoppedWorkerRecord) > envs.minWorkers {
		delete(RunningWorkerRecord, ID)
		StoppedWorkerRecord = append(StoppedWorkerRecord, ID)
		stopContainer(ctx, cli, container)
	}
}

func startContainer(ctx context.Context, cli *client.Client, ID string, envs Envs, updates chan ContainerDetail) {
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

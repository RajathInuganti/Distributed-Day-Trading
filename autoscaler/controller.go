package main

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func createContainers(ctx context.Context, cli *client.Client, networkID string, envs Envs) []string {

	containers := make([]string, envs.maxWorkers)
	for workerNum := 0; workerNum < envs.maxWorkers; workerNum++ {

		hostConfig := &container.HostConfig{
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
		}

		config := &container.Config{
			Image: "txserver",
			Cmd:   []string{"sh", "-c", "/wait && /src/main"},
			Env: []string{
				"MONGODB_URI=mongodb://mongodb:27017/?maxPoolSize=20&w=majority",
				"WAIT_HOSTS=rabbitmq:5672, mongodb:27017",
				"WAIT_HOSTS_TIMEOUT=5",
				"WAIT_SLEEP_INTERVAL=5",
				"WAIT_HOST_CONNECT_TIMEOUT=5",
				"WAIT_BEFORE_HOSTS=5",
			},
		}

		container, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "txserver_"+strconv.Itoa(workerNum))
		if err != nil {
			panic(err)
		}
		err = cli.NetworkConnect(ctx, networkID, container.ID, &network.EndpointSettings{})
		if err != nil {
			panic(err)
		}

		containers[workerNum] = container.ID
	}
	return containers
}

func startContainer(ctx context.Context, cli *client.Client, containerList []string, envs Envs, updates chan string) []string {

	if len(containerList) == 0 {
		return containerList
	}

	containerID, containerList := containerList[0], containerList[1:]
	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	go monitor(ctx, cli, containerID, envs, updates)
	return containerList
}

func stopContainer(ctx context.Context, cli *client.Client, ID string) {
	if err := cli.ContainerStop(ctx, ID, nil); err != nil {
		panic(err)
	}
}

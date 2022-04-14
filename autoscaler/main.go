package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {

	envs := Envs{}
	setup(&envs)

	var networkID string
	updates := make(chan string)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Duration(envs.wait) * time.Second)
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	networks, _ := cli.NetworkList(ctx, types.NetworkListOptions{})
	for _, network := range networks {
		if strings.Contains(network.Name, "txnetwork") {
			networkID = network.ID
			break
		}
	}

	containerList := createContainers(ctx, cli, networkID, envs)

	for _, container := range containers {
		if container.Image == "txserver" {
			go monitor(ctx, cli, container.ID, envs, updates)
		}
	}

	for {
		select {
		case container := <-updates:
			log.Printf("CPU Threshold exceeded for: %s", container)
			containerList = startContainer(ctx, cli, containerList, envs, updates)
		default:
			continue
		}
	}
}

func setup(envs *Envs) {

	envMap := make(map[string]int)

	for _, env := range os.Environ() {

		envPair := strings.Split(env, "=")
		value, _ := strconv.Atoi(envPair[1])
		envMap[envPair[0]] = value

	}

	envs.wait = envMap["WAIT_PERIOD"]
	envs.period = envMap["AUTOSCALER_CHECK_PERIOD"]
	envs.cpuUpper = envMap["CPU_UPPER_THRESHOLD"]
	envs.maxWorkers = envMap["MAX_WORKERS"]

}

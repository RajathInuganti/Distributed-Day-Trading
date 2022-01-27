package main

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var RunningWorkerRecord map[string]bool
var StoppedWorkerRecord []string

func main() {

	envs := Envs{}
	setup(&envs)

	RunningWorkerRecord = make(map[string]bool, envs.maxWorkers)
	StoppedWorkerRecord = make([]string, envs.maxWorkers)
	updates := make(chan ContainerDetail)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Call the createContainers function here and remove the ContainerList query

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		go monitor(ctx, cli, container.ID, envs, updates)
	}

	for {
		update := <-updates
		updateWorkerRecords(ctx, cli, envs, update.containerName, update.startContainer, updates)
	}
}

func setup(envs *Envs) {

	envMap := make(map[string]int)

	for _, env := range os.Environ() {

		envPair := strings.Split(env, "=")
		value, _ := strconv.Atoi(envPair[1])
		envMap[envPair[0]] = value

	}

	envs.cpu = envMap["CPU_ALLOCATION"]
	envs.period = envMap["AUTOSCALER_CHECK_PERIOD"]
	envs.cpuLower = envMap["CPU_LOWER_THRESHOLD"]
	envs.cpuUpper = envMap["CPU_UPPER_THRESHOLD"]
	envs.maxWorkers = envMap["MAX_WORKERS"]
	envs.minWorkers = envMap["MIN_WORKERS"]

}

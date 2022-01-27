package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var RunningWorkerRecord []string
var StoppedWorkerRecord []string

func updateWorkerRecords(activate bool) string {

	var container string

	if activate {
		container, StoppedWorkerRecord = StoppedWorkerRecord[0], StoppedWorkerRecord[1:]
		RunningWorkerRecord = append(RunningWorkerRecord, container)
		return container
	}

	container, RunningWorkerRecord = RunningWorkerRecord[0], RunningWorkerRecord[1:]
	StoppedWorkerRecord = append(StoppedWorkerRecord, container)
	return container
}

func createContainers(ctx context.Context, cli *client.Client, envs Envs) []string {

	containers := make([]string, envs.maxWorkers)
	for i := 0; i < envs.maxWorkers; i++ {
		// Create the maximum number of containers allowed. Start and stop them as necessary
		// Yet to be implemented as the transaction worker needs to be implemented as well
	}
	return containers
}

func main() {

	envs := Envs{}
	setup(&envs)
	RunningWorkerRecord = make([]string, envs.maxWorkers)
	StoppedWorkerRecord = make([]string, envs.maxWorkers)

	var wg sync.WaitGroup
	var m sync.Mutex

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
		wg.Add(envs.minWorkers)
		go monitor(ctx, cli, container.ID, envs, &wg, &m)
	}

	wg.Wait()
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

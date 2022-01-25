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

var WorkerRecord map[string]bool

func updateRecord(containerID string, m *sync.Mutex, active bool) {
	m.Lock()
	WorkerRecord[containerID] = active
	m.Unlock()
}

func main() {

	envs := Envs{}
	setup(&envs)
	WorkerRecord = make(map[string]bool)

	var wg sync.WaitGroup
	var m sync.Mutex

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		updateRecord(container.ID, &m, true)
		wg.Add(1)
		go monitor(ctx, cli, container.ID, envs, &wg)
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

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

	updates := make(chan string)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Call the createContainers function here and remove the ContainerList query
	time.Sleep(time.Duration(envs.period) * 5 * time.Second)
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		go monitor(ctx, cli, container.ID, envs, updates)
	}

	for {
		select {
		case container := <-updates:
			log.Printf("CPU Threshold exceeded for: %s", container)
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

	envs.period = envMap["AUTOSCALER_CHECK_PERIOD"]
	envs.cpuUpper = envMap["CPU_UPPER_THRESHOLD"]
	envs.maxWorkers = envMap["MAX_WORKERS"]

}

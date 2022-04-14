package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func monitor(ctx context.Context, cli *client.Client, ID string, envs Envs, updates chan string) {

	var name string
	stats := new(DockerContainerStats)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if container.ID == ID {
			name = container.Names[0]
		}
	}

	for range time.NewTicker(time.Duration(envs.period) * time.Second).C {

		containerStats, err := cli.ContainerStats(ctx, ID, false)
		if err != nil {
			panic(err)
		}

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(containerStats.Body); err != nil {
			panic(err)
		}
		bytes := buf.Bytes()
		if err := json.Unmarshal(bytes, &stats); err != nil {
			panic(err)
		}

		cpuDelta := stats.CPUStats.CPUUsage.TotalUsage - stats.PrecpuStats.CPUUsage.TotalUsage
		systemCpuDelta := stats.CPUStats.SystemCPUUsage - stats.PrecpuStats.SystemCPUUsage
		numberCpus := float32(stats.CPUStats.OnlineCpus)
		CPUUsage := (float32(cpuDelta) / float32(systemCpuDelta)) * numberCpus * 100.0

		if CPUUsage > float32(envs.cpuUpper) {
			log.Printf("CPU Usage for %s: %f", name, CPUUsage)
			updates <- name
		}
	}
}

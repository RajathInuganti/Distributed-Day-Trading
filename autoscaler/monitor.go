package main

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/docker/docker/client"
)

func monitor(ctx context.Context, cli *client.Client, ID string, envs Envs, updates chan ContainerDetail) {

	stats := new(DockerContainerStats)

	for range time.NewTicker(time.Duration(envs.period) * time.Second).C {

		if _, ok := RunningWorkerRecord[ID]; !ok {
			break
		}

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

		CPUUsageActual := (CPUUsage / float32(envs.cpu)) * 100.0

		if CPUUsageActual > float32(envs.cpuUpper) {

			var update ContainerDetail
			update.startContainer = true
			updates <- update
			continue

		}

		if CPUUsageActual < float32(envs.cpuLower) {

			var update ContainerDetail
			update.startContainer = false
			update.containerName = ID
			updates <- update
			continue

		}
	}
}

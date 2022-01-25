package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/client"
)

func monitor(ctx context.Context, cli *client.Client, ID string, envs Envs, wg *sync.WaitGroup) {

	stats := new(DockerContainerStats)

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

		if CPUUsageActual := (CPUUsage / float32(envs.cpu)) * 100.0; CPUUsageActual > float32(envs.cpuUpper) {

			if len(WorkerRecord) < envs.maxWorkers {
				log.Printf("CPU Threshold exceeded. Starting up new worker.")
			}
			continue
		}

		log.Printf("CPU Utilization at normal levels.")

	}

	wg.Done()
}

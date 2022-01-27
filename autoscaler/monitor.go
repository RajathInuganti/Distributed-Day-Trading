package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func updateStoppedOrRunningContainer(ctx context.Context, cli *client.Client, envs Envs, wg *sync.WaitGroup, m *sync.Mutex, activate bool) {

	m.Lock()
	if activate {

		if len(RunningWorkerRecord) < envs.maxWorkers {
			log.Printf("CPU Threshold exceeded. Starting up worker.")

			containerID := updateWorkerRecords(true)
			if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
				panic(err)
			}
			wg.Add(envs.minWorkers)
			go monitor(ctx, cli, containerID, envs, wg, m)
		}
		return
	}

	if len(RunningWorkerRecord) > envs.minWorkers {
		log.Printf("CPU Utilization at below normal containers. Stopping worker.")

		containerID := updateWorkerRecords(false)
		if err := cli.ContainerStop(ctx, containerID, nil); err != nil {
			panic(err)
		}
	}
	m.Unlock()
}

func monitor(ctx context.Context, cli *client.Client, ID string, envs Envs, wg *sync.WaitGroup, m *sync.Mutex) {

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

		if cpuDelta == 0 {
			break
		}

		CPUUsageActual := (CPUUsage / float32(envs.cpu)) * 100.0
		if CPUUsageActual > float32(envs.cpuUpper) {
			updateStoppedOrRunningContainer(ctx, cli, envs, wg, m, true)
			continue
		}

		if CPUUsageActual < float32(envs.cpuLower) {
			updateStoppedOrRunningContainer(ctx, cli, envs, wg, m, false)
			continue
		}
	}

	wg.Done()
}

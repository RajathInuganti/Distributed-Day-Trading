package main

import "time"

type Envs struct {
	wait       int
	period     int
	cpuUpper   int
	maxWorkers int
}

type DockerContainerStats struct {
	Read      time.Time `json:"read"`
	Preread   time.Time `json:"preread"`
	PidsStats struct {
		Current int    `json:"current"`
		Limit   uint64 `json:"limit"`
	} `json:"pids_stats"`
	BlkioStats struct {
		IoServiceBytesRecursive interface{} `json:"io_service_bytes_recursive"`
		IoServicedRecursive     interface{} `json:"io_serviced_recursive"`
		IoQueueRecursive        interface{} `json:"io_queue_recursive"`
		IoServiceTimeRecursive  interface{} `json:"io_service_time_recursive"`
		IoWaitTimeRecursive     interface{} `json:"io_wait_time_recursive"`
		IoMergedRecursive       interface{} `json:"io_merged_recursive"`
		IoTimeRecursive         interface{} `json:"io_time_recursive"`
		SectorsRecursive        interface{} `json:"sectors_recursive"`
	} `json:"blkio_stats"`
	NumProcs     int `json:"num_procs"`
	StorageStats struct {
	} `json:"storage_stats"`
	CPUStats struct {
		CPUUsage struct {
			TotalUsage        uint64 `json:"total_usage"`
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCpus     int    `json:"online_cpus"`
		ThrottlingData struct {
			Periods          uint64 `json:"periods"`
			ThrottledPeriods uint64 `json:"throttled_periods"`
			ThrottledTime    uint64 `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"cpu_stats"`
	PrecpuStats struct {
		CPUUsage struct {
			TotalUsage        uint64 `json:"total_usage"`
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCpus     int    `json:"online_cpus"`
		ThrottlingData struct {
			Periods          uint64 `json:"periods"`
			ThrottledPeriods uint64 `json:"throttled_periods"`
			ThrottledTime    uint64 `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Stats struct {
			ActiveAnon            uint64 `json:"active_anon"`
			ActiveFile            uint64 `json:"active_file"`
			Anon                  uint64 `json:"anon"`
			AnonThp               uint64 `json:"anon_thp"`
			File                  uint64 `json:"file"`
			FileDirty             uint64 `json:"file_dirty"`
			FileMapped            uint64 `json:"file_mapped"`
			FileWriteback         uint64 `json:"file_writeback"`
			InactiveAnon          uint64 `json:"inactive_anon"`
			InactiveFile          uint64 `json:"inactive_file"`
			KernelStack           uint64 `json:"kernel_stack"`
			Pgactivate            uint64 `json:"pgactivate"`
			Pgdeactivate          uint64 `json:"pgdeactivate"`
			Pgfault               uint64 `json:"pgfault"`
			Pglazyfree            uint64 `json:"pglazyfree"`
			Pglazyfreed           uint64 `json:"pglazyfreed"`
			Pgmajfault            uint64 `json:"pgmajfault"`
			Pgrefill              uint64 `json:"pgrefill"`
			Pgscan                uint64 `json:"pgscan"`
			Pgsteal               uint64 `json:"pgsteal"`
			Shmem                 uint64 `json:"shmem"`
			Slab                  uint64 `json:"slab"`
			SlabReclaimable       uint64 `json:"slab_reclaimable"`
			SlabUnreclaimable     uint64 `json:"slab_unreclaimable"`
			Sock                  uint64 `json:"sock"`
			ThpCollapseAlloc      uint64 `json:"thp_collapse_alloc"`
			ThpFaultAlloc         uint64 `json:"thp_fault_alloc"`
			Unevictable           uint64 `json:"unevictable"`
			WorkingsetActivate    uint64 `json:"workingset_activate"`
			WorkingsetNodereclaim uint64 `json:"workingset_nodereclaim"`
			WorkingsetRefault     uint64 `json:"workingset_refault"`
		} `json:"stats"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	Name     string `json:"name"`
	ID       string `json:"id"`
	Networks struct {
		Eth0 struct {
			RxBytes   uint64 `json:"rx_bytes"`
			RxPackets uint64 `json:"rx_packets"`
			RxErrors  uint64 `json:"rx_errors"`
			RxDropped uint64 `json:"rx_dropped"`
			TxBytes   uint64 `json:"tx_bytes"`
			TxPackets uint64 `json:"tx_packets"`
			TxErrors  uint64 `json:"tx_errors"`
			TxDropped uint64 `json:"tx_dropped"`
		} `json:"eth0"`
	} `json:"networks"`
}

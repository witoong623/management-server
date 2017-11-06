package main

import (
	"fmt"
)

type ContainerState struct {
	BlkioStats   CtnBlockIO   `json:"blkio_stats"`
	CPUStats     CtnCPU       `json:"cpu_stats"`
	ID           string       `json:"id"`
	MemoryStats  CtnMemory    `json:"memory_stats"`
	Name         string       `json:"name"`
	Networks     CtnNetwork   `json:"networks"`
	NumProcs     int64        `json:"num_procs"`
	PidsStats    CtnProcessID `json:"pids_stats"`
	PrecpuStats  CtnCPU       `json:"precpu_stats"`
	Preread      string       `json:"preread"`
	Read         string       `json:"read"`
	StorageStats CtnStorage   `json:"storage_stats"`
}

func (c *ContainerState) String() string {
	return fmt.Sprintf("CPU percent: %b\n", c.CPUStats.CPUUsage.TotalUsage-c.PrecpuStats.CPUUsage.TotalUsage)
}

type MemoryStats struct {
	ActiveAnon              int64 `json:"active_anon"`
	ActiveFile              int64 `json:"active_file"`
	Cache                   int64 `json:"cache"`
	Dirty                   int64 `json:"dirty"`
	HierarchicalMemoryLimit int64 `json:"hierarchical_memory_limit"`
	HierarchicalMemswLimit  int64 `json:"hierarchical_memsw_limit"`
	InactiveAnon            int64 `json:"inactive_anon"`
	InactiveFile            int64 `json:"inactive_file"`
	MappedFile              int64 `json:"mapped_file"`
	Pgfault                 int64 `json:"pgfault"`
	Pgmajfault              int64 `json:"pgmajfault"`
	Pgpgin                  int64 `json:"pgpgin"`
	Pgpgout                 int64 `json:"pgpgout"`
	Rss                     int64 `json:"rss"`
	RssHuge                 int64 `json:"rss_huge"`
	Swap                    int64 `json:"swap"`
	TotalActiveAnon         int64 `json:"total_active_anon"`
	TotalActiveFile         int64 `json:"total_active_file"`
	TotalCache              int64 `json:"total_cache"`
	TotalDirty              int64 `json:"total_dirty"`
	TotalInactiveAnon       int64 `json:"total_inactive_anon"`
	TotalInactiveFile       int64 `json:"total_inactive_file"`
	TotalMappedFile         int64 `json:"total_mapped_file"`
	TotalPgfault            int64 `json:"total_pgfault"`
	TotalPgmajfault         int64 `json:"total_pgmajfault"`
	TotalPgpgin             int64 `json:"total_pgpgin"`
	TotalPgpgout            int64 `json:"total_pgpgout"`
	TotalRss                int64 `json:"total_rss"`
	TotalRssHuge            int64 `json:"total_rss_huge"`
	TotalSwap               int64 `json:"total_swap"`
	TotalUnevictable        int64 `json:"total_unevictable"`
	TotalWriteback          int64 `json:"total_writeback"`
	Unevictable             int64 `json:"unevictable"`
	Writeback               int64 `json:"writeback"`
}

type CtnCPU struct {
	CPUUsage       CPUUsageStats     `json:"cpu_usage"`
	OnlineCpus     int64             `json:"online_cpus"`
	SystemCPUUsage int64             `json:"system_cpu_usage"`
	ThrottlingData CPUThrottlingData `json:"throttling_data"`
}

type CtnProcessID struct {
	Current int64 `json:"current"`
}

// CtnNetwork contains network information related to eth0 interface.
type CtnNetwork struct {
	DefaultInterface NetworkInterfaceStats `json:"eth0"`
}

type CtnBlockIO struct {
	IoMergedRecursive       []interface{}             `json:"io_merged_recursive"`
	IoQueueRecursive        []interface{}             `json:"io_queue_recursive"`
	IoServiceBytesRecursive []IOServiceBytesRecursive `json:"io_service_bytes_recursive"`
	IoServiceTimeRecursive  []interface{}             `json:"io_service_time_recursive"`
	IoServicedRecursive     []IOServiceBytesRecursive `json:"io_serviced_recursive"`
	IoTimeRecursive         []interface{}             `json:"io_time_recursive"`
	IoWaitTimeRecursive     []interface{}             `json:"io_wait_time_recursive"`
	SectorsRecursive        []interface{}             `json:"sectors_recursive"`
}

type CtnMemory struct {
	Limit    int64       `json:"limit"`
	MaxUsage int64       `json:"max_usage"`
	Stats    MemoryStats `json:"stats"`
	Usage    int64       `json:"usage"`
}

type IOServiceBytesRecursive struct {
	Major int64  `json:"major"`
	Minor int64  `json:"minor"`
	Op    string `json:"op"`
	Value int64  `json:"value"`
}

type CPUUsageStats struct {
	PercpuUsage       []int64 `json:"percpu_usage"`
	TotalUsage        int64   `json:"total_usage"`
	UsageInKernelmode int64   `json:"usage_in_kernelmode"`
	UsageInUsermode   int64   `json:"usage_in_usermode"`
}

type CPUThrottlingData struct {
	Periods          int64 `json:"periods"`
	ThrottledPeriods int64 `json:"throttled_periods"`
	ThrottledTime    int64 `json:"throttled_time"`
}

type NetworkInterfaceStats struct {
	RxBytes   int64 `json:"rx_bytes"`
	RxDropped int64 `json:"rx_dropped"`
	RxErrors  int64 `json:"rx_errors"`
	RxPackets int64 `json:"rx_packets"`
	TxBytes   int64 `json:"tx_bytes"`
	TxDropped int64 `json:"tx_dropped"`
	TxErrors  int64 `json:"tx_errors"`
	TxPackets int64 `json:"tx_packets"`
}

type CtnStorage struct{}

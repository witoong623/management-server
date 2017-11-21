package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	dockerClient "github.com/docker/docker/client"
)

const (
	// DockerVersion is version of Docker API
	DockerVersion = "1.32"
)

// Container represents container that running in specific node.
type Container struct {
	cancelSignal chan struct{}
	ContainerID  string
	ServiceName  string
}

// NodeStats holds information about stats of specific node.
type NodeStats struct {
}

// DockerMonitor holds information necessary for mornitoring.
type DockerMonitor struct {
	httpClient      *http.Client
	edgeNodeCtxs    map[string]nodeMonitoringCtx
	updateStatsChan chan NodeStats
}

type nodeMonitoringCtx struct {
	dockerMonitorCtx     *DockerMonitor
	dockerClient         *dockerClient.Client
	monitoringContainers map[string]Container
	nodeAddr             string
	nodeCancelSignal     chan struct{}
	targetContainer      chan Container
	updateCPUStats       chan int64
}

type containerMonitoringCtx struct {
}

// NewDockerMonitor initiates any necessary type for starting mornitoring resource.
func NewDockerMonitor() *DockerMonitor {
	// Must use zero value of http.Client, it differs from http.Client{}
	var httpClient *http.Client

	return &DockerMonitor{
		httpClient:      httpClient,
		edgeNodeCtxs:    make(map[string]nodeMonitoringCtx),
		updateStatsChan: make(chan NodeStats),
	}
}

// MonitorContainer start monitoring container in specific node.
func (dmCtx *DockerMonitor) MonitorContainer(nodeAddr, conID, serviceName string) {
	// First, check if this node is being monitored or not.
	node, ok := dmCtx.edgeNodeCtxs[nodeAddr]

	if ok {
		// If this node is being monitored, just add new container object that want it to be monitored.
		node.targetContainer <- Container{ServiceName: serviceName, ContainerID: conID}
	} else {
		// If this node haven't been monitored, start monitoring this node.
		// First, create docker client, 1 client per node.
		client, err := dockerClient.NewClient(nodeAddr, DockerVersion, dmCtx.httpClient, nil)
		if err != nil {
			log.Fatalln(err)
		}

		monitorContext := nodeMonitoringCtx{
			dockerMonitorCtx:     dmCtx,
			dockerClient:         client,
			monitoringContainers: make(map[string]Container),
			nodeAddr:             nodeAddr,
			nodeCancelSignal:     make(chan struct{}, 1),
			targetContainer:      make(chan Container),
			updateCPUStats:       make(chan int64),
		}

		container := Container{
			cancelSignal: make(chan struct{}, 1),
			ContainerID:  conID,
			ServiceName:  serviceName,
		}

		dmCtx.edgeNodeCtxs[nodeAddr] = monitorContext
		go monitorContext.startMonitoringNode()
		monitorContext.targetContainer <- container

	}
}

// StopMonitor as the name suggests.
func (dmCtx *DockerMonitor) StopMonitor() {
	for _, nodeCtx := range dmCtx.edgeNodeCtxs {
		nodeCtx.stopMonitoringNode()
	}
}

// StopMonitorNode stops monitor every container in specific node.
func (dmCtx *DockerMonitor) StopMonitorNode(nodeAddr string) error {
	node, ok := dmCtx.edgeNodeCtxs[nodeAddr]
	if ok {
		node.stopMonitoringNode()
		return nil
	}
	return fmt.Errorf("node %s not found", nodeAddr)
}

func (nCtx *nodeMonitoringCtx) startMonitoringNode() {
	stopReceiveUpdateSignal := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case container := <-nCtx.targetContainer:
				go nCtx.queryStatus(container.ContainerID, container.cancelSignal)
				nCtx.monitoringContainers[container.ContainerID] = container
			case <-nCtx.nodeCancelSignal:
				for _, container := range nCtx.monitoringContainers {
					container.cancelSignal <- struct{}{}
				}
				stopReceiveUpdateSignal <- struct{}{}
				break
			}
		}
	}()

	for {
		select {
		case cpuPercent := <-nCtx.updateCPUStats:
			fmt.Printf("Node Addr: %s, CPU usage: %d%%\n", nCtx.nodeAddr, cpuPercent)
		case <-stopReceiveUpdateSignal:
			break
		}
	}
}

func (nCtx *nodeMonitoringCtx) stopMonitoringNode() {
	nCtx.nodeCancelSignal <- struct{}{}
}

func (nCtx *nodeMonitoringCtx) queryStatus(containerID string, stopSignal <-chan struct{}) {
	stats, err := nCtx.dockerClient.ContainerStats(context.Background(), containerID, true)
	if err != nil {
		log.Printf("Error getting stats: %s", err)
		return
	}
	defer stats.Body.Close()

	decoder := json.NewDecoder(stats.Body)
	var unused ContainerStats
	if err := decoder.Decode(&unused); err != nil {
		log.Println("Error from first decode.")
		return
	}
	for {
		select {
		case <-stopSignal:
			return
		default:
			var containerStats ContainerStats
			if err := decoder.Decode(&containerStats); err != nil {
				if err == io.EOF {
					log.Println("Remote Docker closed connection.")
					return
				}
				log.Println(err)
				return
			}
			cpuPercentUsage := (containerStats.CPUStats.CPUUsage.TotalUsage - containerStats.PrecpuStats.CPUUsage.TotalUsage) / 10000000
			nCtx.updateCPUStats <- cpuPercentUsage
		}
	}
}

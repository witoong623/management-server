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
func (dmCtx *DockerMonitor) MonitorContainer(nodeAddr, conID, serviceName string) error {
	// First, check if this node is being monitored or not.
	node, ok := dmCtx.edgeNodeCtxs[nodeAddr]

	if ok {
		// If this node is being monitored, just add new container object that want it to be monitored.
		node.targetContainer <- Container{ServiceName: serviceName, ContainerID: conID, cancelSignal: make(chan struct{}, 1)}
	} else {
		// If this node haven't been monitored, start monitoring this node.
		// First, create docker client, 1 client per node.
		client, err := dockerClient.NewClient(nodeAddr, DockerVersion, dmCtx.httpClient, nil)
		if err != nil {
			return err
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
	return nil
}

// StopMonitor as the name suggests.
func (dmCtx *DockerMonitor) StopMonitor() {
	for _, nodeCtx := range dmCtx.edgeNodeCtxs {
		nodeCtx.stopMonitoringNode()
	}
	dmCtx.edgeNodeCtxs = make(map[string]nodeMonitoringCtx)
}

// StopMonitorNode stops monitor every container in specific node.
func (dmCtx *DockerMonitor) StopMonitorNode(nodeAddr string) error {
	node, ok := dmCtx.edgeNodeCtxs[nodeAddr]
	if ok {
		node.stopMonitoringNode()
		delete(dmCtx.edgeNodeCtxs, nodeAddr)
		return nil
	}
	return fmt.Errorf("node %s not found", nodeAddr)
}

// StopMonitorContainer stops monitor specific container in specfici node.
func (dmCtx *DockerMonitor) StopMonitorContainer(nodeAddr, conID string) error {
	node, ok := dmCtx.edgeNodeCtxs[nodeAddr]
	if ok {
		container, ok := node.monitoringContainers[conID]
		if ok {
			container.cancelSignal <- struct{}{}
			delete(node.monitoringContainers, conID)
			if len(node.monitoringContainers) == 0 {
				node.stopMonitoringNode()
				delete(dmCtx.edgeNodeCtxs, nodeAddr)
			}
			return nil
		}
		return fmt.Errorf("container ID %s not found or not being monitored", conID)
	}
	return fmt.Errorf("node address %s not found or not being monitored", nodeAddr)
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

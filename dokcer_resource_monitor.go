package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	dockerClient "github.com/docker/docker/client"
)

const (
	// DockerVersion is version of Docker API
	DockerVersion = "1.32"
)

// Container represents container that running in specific node.
type Container struct {
	ContainerID string
	ServiceName string
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

// nodeMonitoringCtx holds information related to specific node when monitoring.
type nodeMonitoringCtx struct {
	dockerMonitorCtx    *DockerMonitor
	dockerClient        *dockerClient.Client
	httpClient          *http.Client
	monitoredContainers map[string]Container
	nodeAddr            string
	nodeCancelSignal    chan struct{}
	ticker              *time.Ticker
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
		node.monitoredContainers[conID] = Container{ServiceName: serviceName, ContainerID: conID}
	} else {
		// If this node haven't been monitored, start monitoring this node.
		// First, create docker client, 1 client per node.
		client, err := dockerClient.NewClient(nodeAddr, DockerVersion, dmCtx.httpClient, nil)
		if err != nil {
			log.Fatalln(err)
		}

		monitorContext := nodeMonitoringCtx{
			dockerMonitorCtx:    dmCtx,
			dockerClient:        client,
			monitoredContainers: make(map[string]Container),
			nodeAddr:            nodeAddr,
			nodeCancelSignal:    make(chan struct{}),
			ticker:              time.NewTicker(time.Second * 10),
		}

		monitorContext.monitoredContainers[conID] = Container{ServiceName: serviceName, ContainerID: conID}
		dmCtx.edgeNodeCtxs[nodeAddr] = monitorContext

		go monitorContext.startMonitoringNode()
	}
}

func (nCtx *nodeMonitoringCtx) startMonitoringNode() {
	for {
		select {
		case <-nCtx.ticker.C:
			nCtx.queryStatus()
		case <-nCtx.nodeCancelSignal:
			break
		}
	}
}

func (nCtx *nodeMonitoringCtx) stopMonitoringNode() {
	nCtx.ticker.Stop()
	nCtx.nodeCancelSignal <- struct{}{}
}

func (nCtx *nodeMonitoringCtx) queryStatus() {
	for _, container := range nCtx.monitoredContainers {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		stats, err := nCtx.dockerClient.ContainerStats(ctx, container.ContainerID, false)
		if err != nil {
			log.Fatalf("Error getting stats: %s", err)
			continue
		}

		defer stats.Body.Close()
		msgBytes, _ := ioutil.ReadAll(stats.Body)
		var containerStats ContainerState
		err = json.Unmarshal(msgBytes, &containerStats)
		if err != nil {
			log.Printf("JSON Container stats decode error. %s\n", err)
		}
		fmt.Println(&containerStats)
	}
}

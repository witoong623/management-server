package main

import (
	"log"
	"sync"
)

// CloudletNode encapsulates information about Cloudlet
type CloudletNode struct {
	// Name of cloudlet is IPv4 address, same as Address field
	Name string

	// IPAddr is IPv4 address of this cloudlet node in network
	IPAddr string

	// DomainName is domain name that point to IPAddr, in case it is available.
	DomainName string

	// AvailableServices is the collection of service name that are currently
	// available in this Cloudlet.
	AvailableServices []ServiceInfo

	cloudletQueryService *CloudletQueryService
	workloadMutex        sync.Mutex
	currentWorkload      int32
}

// NewCloudletNode is a factory function that creates new instance of CloudletNode
// and setup any necessary service for this Cloudlet i.e. monitoring server.
func NewCloudletNode(name, ip, domain string) (*CloudletNode, error) {
	queryService := NewCloudletQueryService()
	cloudlet := &CloudletNode{
		Name:                 name,
		IPAddr:               ip,
		cloudletQueryService: queryService,
	}

	log.Printf("got register request from %v, IP %v and domain %v", name, ip, domain)
	statusChan := queryService.QueryCloudletWorkload(cloudlet)

	// TODO disable for testing
	go cloudlet.updateCloudletWorkload(statusChan)

	return cloudlet, nil
}

// ServiceInfo encapsulates data about a service
type ServiceInfo struct {
	Name string
}

// GetCurrentWorkload returns current workload of this Cloudlet.
func (c *CloudletNode) GetCurrentWorkload() int32 {
	c.workloadMutex.Lock()
	defer c.workloadMutex.Unlock()
	return c.currentWorkload
}

func (c *CloudletNode) updateCloudletWorkload(workloadChan <-chan WorkloadStatusMessage) {
	for status := range workloadChan {
		c.workloadMutex.Lock()
		c.currentWorkload = status.ClientCount
		log.Printf("Current workload is %s", c.currentWorkload)
		c.workloadMutex.Unlock()
	}
}

package main

import (
	"context"
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
	AvailableServices []Service

	cloudletQueryService *CloudletQueryService
	workloadMutex        sync.Mutex
	currentWorkload      int32

	cancelWorkloadQuery context.CancelFunc
}

// NewCloudletNode is a factory function that creates new instance of CloudletNode
// and setup any necessary service for this Cloudlet i.e. monitoring server.
func NewCloudletNode(name, ip, domain string) (*CloudletNode, error) {
	queryService := NewCloudletQueryService()
	ctx, cancelWorkloadQuery := context.WithCancel(context.Background())
	cloudlet := &CloudletNode{
		Name:                 name,
		IPAddr:               ip,
		cloudletQueryService: queryService,
		cancelWorkloadQuery:  cancelWorkloadQuery,
	}

	log.Printf("got register request from %v, IP %v and domain %v\n", name, ip, domain)
	statusChan := queryService.QueryCloudletWorkload(ctx, cloudlet)
	go cloudlet.updateCloudletWorkload(statusChan)

	return cloudlet, nil
}

// GetCurrentWorkload returns current workload of this Cloudlet.
func (c *CloudletNode) GetCurrentWorkload() int32 {
	c.workloadMutex.Lock()
	defer c.workloadMutex.Unlock()
	return c.currentWorkload
}

func (c *CloudletNode) updateCloudletWorkload(workloadChan <-chan WorkloadStatusMessage) {
	more := true
	for status := range workloadChan {
		c.workloadMutex.Lock()
		c.currentWorkload = status.ClientCount
		if c.currentWorkload > 0 || more {
			// TODO: This is for debug only, delete it when do system testing
			log.Printf("Current workload is %d", c.currentWorkload)
			more = !more
		}
		c.workloadMutex.Unlock()
	}
}

// UnRegisterCloudlet stops any operation with given Cloudlet
func (c *CloudletNode) UnRegisterCloudlet() {
	c.cancelWorkloadQuery()
}

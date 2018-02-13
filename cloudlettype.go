package main

import (
	"context"
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
	AvailableServices []*Service

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
		AvailableServices:    make([]*Service, 0, 10),
	}

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

// SetCurrentWorkload sets current workload of cloudlet
func (c *CloudletNode) SetCurrentWorkload(val int32) {
	c.workloadMutex.Lock()
	defer c.workloadMutex.Unlock()
	c.currentWorkload = val
}

func (c *CloudletNode) updateCloudletWorkload(workloadChan <-chan WorkloadStatusMessage) {
	for status := range workloadChan {
		c.workloadMutex.Lock()
		c.currentWorkload = status.ClientCount
		c.workloadMutex.Unlock()
	}
}

// UnRegisterCloudlet stops any operation with given Cloudlet
func (c *CloudletNode) UnRegisterCloudlet() {
	// calling this method will signal cancel context, workload query service closes outChan
	// updateCloudletWorkload stop working (because channel is close)
	c.cancelWorkloadQuery()
}

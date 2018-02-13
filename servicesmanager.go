package main

import (
	"errors"
	"sync"
)

type ServiceManager struct {
	serMutex          sync.RWMutex
	availableServices map[string]*Service
}

// Service holds information about service that is available in edge network
// it doesn't hold information about which node have this service
type Service struct {
	Name   string
	Domain string
}

// NewServiceManager news and returns instance of ServiceManager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		availableServices: make(map[string]*Service),
	}
}

// GetService returns service associated with domain name
func (s *ServiceManager) GetService(domain string) (*Service, error) {
	s.serMutex.RLock()
	service, ok := s.availableServices[domain]
	s.serMutex.RUnlock()
	if !ok {
		return nil, errors.New("service not found")
	}
	return service, nil
}

// AddService adds service to service manager.
func (s *ServiceManager) AddService(domain string, service *Service) {
	s.serMutex.Lock()
	s.availableServices[domain] = service
	s.serMutex.Unlock()
}

package main

import (
	"errors"
	"sync"
)

type ServiceManager struct {
	serMutex          sync.RWMutex
	availableServices map[string]Service
}

type Service struct {
	Name   string
	Domain string
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		availableServices: make(map[string]Service),
	}
}

func (s *ServiceManager) GetService(domain string) (*Service, error) {
	s.serMutex.RLock()
	service, ok := s.availableServices[domain]
	s.serMutex.RUnlock()
	if !ok {
		return nil, errors.New("service not found")
	}
	return &service, nil
}

func (s *ServiceManager) AddService(domain string, service Service) {
	s.serMutex.Lock()
	s.availableServices[domain] = service
	s.serMutex.Unlock()
}

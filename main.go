package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/witoong623/restserver"
)

type manageCtx struct {
	Cloudlets map[string]*CloudletNode

	ServiceManager *ServiceManager
}

func main() {
	if err := ReadConfigurationFile("config.json"); err != nil {
		log.Fatal(err)
	}
	// parse flags
	servicesFilename := flag.String("service-filename", "", "name of file contains list of domain names")
	flag.Parse()

	ctx := &manageCtx{
		Cloudlets:      make(map[string]*CloudletNode),
		ServiceManager: NewServiceManager(),
	}

	managementRESTServer := restserver.NewRESTServer(":8000")
	managementRESTServer.Handle("/cloudlet/register", HandleCloudletRegisterCommand(ctx))

	go func() {
		log.Println("start REST API server")
		managementRESTServer.StartListening()
	}()

	dnsserver := NewDNSServer(ctx)

	go func() {
		dnsserver.ListenAndServe()
	}()

	// TODO: Testing purpose code, must be delete if we actually deploy
	serviceList, err := buildTestServicesFromFile(*servicesFilename)
	if err != nil {
		log.Fatal(err)
	}

	for _, service := range serviceList {
		ctx.ServiceManager.AddService(service.Domain, service)
	}

	_, cancel1 := context.WithCancel(context.Background())
	cloudlet1 := &CloudletNode{
		Name:                 "c1",
		IPAddr:               "172.16.1.100",
		cloudletQueryService: NewCloudletQueryService(),
		cancelWorkloadQuery:  cancel1,
	}

	for _, service := range serviceList {
		cloudlet1.AvailableServices = append(cloudlet1.AvailableServices, service)
	}

	_, cancel2 := context.WithCancel(context.Background())
	cloudlet2 := &CloudletNode{
		Name:                 "c2",
		IPAddr:               "172.16.1.101",
		cloudletQueryService: NewCloudletQueryService(),
		cancelWorkloadQuery:  cancel2,
	}

	for _, service := range serviceList {
		cloudlet2.AvailableServices = append(cloudlet2.AvailableServices, service)
	}

	ctx.Cloudlets[cloudlet1.IPAddr] = cloudlet1
	ctx.Cloudlets[cloudlet2.IPAddr] = cloudlet2

	// Print all of information to confirm debug environment
	log.Printf("there are %v services available in this Edge Network\n", len(ctx.ServiceManager.availableServices))
	for domain, service := range ctx.ServiceManager.availableServices {
		log.Printf("domain \"%v\" has service name %v\n", domain, service.Name)
	}
	log.Printf("there are %v Cloudlet Nodes\n", len(ctx.Cloudlets))
	for cIP, cloudlet := range ctx.Cloudlets {
		log.Printf("cloudlet IP %v have %v services\n", cIP, len(cloudlet.AvailableServices))
		for _, service := range cloudlet.AvailableServices {
			log.Printf("name %v domain %v\n", service.Name, service.Domain)
		}
	}

	// change the number of workloads each cloudlet have every 2 seconds
	randomCancel := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second * 1)

		for {
			select {
			case <-ticker.C:
				cloudlet1.workloadMutex.Lock()
				cloudlet1.currentWorkload = rand.Int31n(5000)
				cloudlet1.workloadMutex.Unlock()

				cloudlet2.workloadMutex.Lock()
				cloudlet2.currentWorkload = rand.Int31n(5000)
				cloudlet2.workloadMutex.Unlock()
			case <-randomCancel:
				ticker.Stop()
				return
			}
		}
	}()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan
	managementRESTServer.StopListening(nil)
	dnsserver.Shutdown()
	for _, cloudlet := range ctx.Cloudlets {
		cloudlet.UnRegisterCloudlet()
	}

	randomCancel <- struct{}{}

	log.Println("Terminating application.")
}

func buildTestServicesFromFile(filename string) ([]Service, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	serviceList := make([]Service, 0, 10)
	for scanner.Scan() {
		domain := scanner.Text()
		splited := strings.Split(domain, ".")
		serviceList = append(serviceList, Service{Name: splited[0], Domain: domain})
	}
	return serviceList, nil
}

// WriteJSONHTTPResponse is a helper function use to write JSON HTTP response.
func WriteJSONHTTPResponse(w http.ResponseWriter, jsonObject interface{}, httpCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(jsonObject)
}

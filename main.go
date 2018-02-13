package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/witoong623/restserver"
)

const (
	// MaxCloudletWorkload is the maximum number of requests that each cloudlet can process at the same time
	MaxCloudletWorkload = 5
)

type manageCtx struct {
	Cloudlets map[string]*CloudletNode

	ServiceManager *ServiceManager
}

func main() {
	if err := ReadConfigurationFile("config.json"); err != nil {
		log.Fatal(err)
	}

	ctx := &manageCtx{
		Cloudlets:      make(map[string]*CloudletNode),
		ServiceManager: NewServiceManager(),
	}

	managementRESTServer := restserver.NewRESTServer(":8000")
	managementRESTServer.Handle("/cloudlet/register", HandleCloudletRegisterCommand(ctx))
	managementRESTServer.Handle("/service/register", HandleServiceRegisterCommand(ctx))

	go func() {
		log.Println("start REST API server")
		managementRESTServer.StartListening()
	}()

	dnsserver := NewDNSServer(ctx)

	go func() {
		dnsserver.ListenAndServe()
	}()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan
	managementRESTServer.StopListening(nil)
	dnsserver.Shutdown()
	for _, cloudlet := range ctx.Cloudlets {
		cloudlet.UnRegisterCloudlet()
	}

	log.Println("Terminating application.")
}

// WriteJSONHTTPResponse is a helper function use to write JSON HTTP response.
func WriteJSONHTTPResponse(w http.ResponseWriter, jsonObject interface{}, httpCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(jsonObject)
}

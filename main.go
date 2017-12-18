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

type edgeProxyCtx struct {
	Cloudlets map[string]CloudletNode
}

func main() {
	ctx := &edgeProxyCtx{}
	managementRESTServer := restserver.NewRESTServer(":8000")
	managementRESTServer.Handle("/cloudlet/register", HandleCloudletRegisterCommand(ctx))

	go func() {
		log.Println("start REST API server")
		managementRESTServer.StartListening()
	}()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan
	managementRESTServer.StopListening(nil)

	log.Println("Terminating application.")
}

// WriteJSONHTTPResponse is a helper function use to write JSON HTTP response.
func WriteJSONHTTPResponse(w http.ResponseWriter, jsonObject interface{}, httpCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(jsonObject)
}

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
)

type edgeNodeCtx struct {
	Services map[string]serviceInfo
}

type serviceInfo struct {
	Name  string
	Nodes []nodeInfo
}

type nodeInfo struct {
	IPAddr     string
	DockerAddr string
}

type queryNodeReturn struct {
	IP string
}

func main() {
	srv := initHTTPServer()
	dockerMonitor := NewDockerMonitor()

	var containerID string
	flag.StringVar(&containerID, "conid", "", "ID of container that is running.")
	flag.Parse()
	if containerID == "" {
		log.Fatalln("Container ID is required.")
	}
	dockerMonitor.MonitorContainer(client.DefaultDockerHost, containerID, "ocr")

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan
	dockerMonitor.StopMonitor()

	log.Println("Terminating application.")
	srv.Shutdown(nil)

}

func initHTTPServer() *http.Server {
	srv := &http.Server{Addr: ":8000"}
	ctx := &edgeNodeCtx{}
	http.Handle("/getnode", HandleServerQuery(ctx))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("ListenAndServ error: %s\n", err)
		}
	}()

	return srv
}

// HandleServerQuery handles available Edge Node query
// It shuold return address of Node server that have less work to process or return 0.0.0.0
// To indicate that mobile devices should connect to cloud server.
func HandleServerQuery(c *edgeNodeCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getQuery := r.URL.Query()
		serviceName := getQuery.Get("service")
		nodeAddress := c.GetComputeNode(serviceName)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(queryNodeReturn{IP: nodeAddress})
	})
}

// GetComputeNode returns IP address of available Node that can serve that service.
// Initiate that service on demand if no service available.
func (c *edgeNodeCtx) GetComputeNode(service string) string {
	log.Printf("Request for service : %s\n", service)
	return "0.0.0.0:0"
}

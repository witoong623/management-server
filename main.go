package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type edgeProxyCtx struct {
	Services      map[string]serviceInfo
	dockerMonitor *DockerMonitor
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
	dockerMonitor := NewDockerMonitor()
	edgeCtx := &edgeProxyCtx{dockerMonitor: dockerMonitor}
	srv := initHTTPServer(edgeCtx)

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan
	dockerMonitor.StopMonitor()

	log.Println("Terminating application.")
	srv.Shutdown(nil)

}

func initHTTPServer(ctx *edgeProxyCtx) *http.Server {
	srv := &http.Server{Addr: ":8000"}
	http.Handle("/getnode", HandleServerQuery(ctx))
	http.Handle("/startmonitor", HandleStartMonitorCommand(ctx))
	http.Handle("/stopmonitor", HandleStopMonitorCommand(ctx))

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
func HandleServerQuery(c *edgeProxyCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		getQuery := r.URL.Query()
		serviceName := getQuery.Get("service")
		nodeAddress := c.GetComputeNode(serviceName)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(queryNodeReturn{IP: nodeAddress})
	})
}

// GetComputeNode returns IP address of available Node that can serve that service.
// Initiate that service on demand if no service available.
func (c *edgeProxyCtx) GetComputeNode(service string) string {
	log.Printf("Request for service : %s\n", service)
	return "0.0.0.0:0"
}

// HandleStartMonitorCommand handles monitor request.
func HandleStartMonitorCommand(c *edgeProxyCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.ParseForm()
		remoteAddr := r.Form.Get("remote-addr")
		containerID := r.Form.Get("conid")
		serviceName := r.Form.Get("service")
		c.dockerMonitor.MonitorContainer(remoteAddr, containerID, serviceName)
		w.Write([]byte(fmt.Sprintf("Started monitoring %s container in node %s", containerID, remoteAddr)))
	})
}

// HandleStopMonitorCommand handles stop monitor request.
func HandleStopMonitorCommand(c *edgeProxyCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.ParseForm()
		remoteAddr := r.Form.Get("remote-addr")
		containerID := r.Form.Get("conid")

		if err := c.dockerMonitor.StopMonitorContainer(remoteAddr, containerID); err != nil {
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte(fmt.Sprintf("Stop monitoring %s container in node %s", containerID, remoteAddr)))
	})
}

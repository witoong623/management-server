package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

var httpClient *http.Client

func init() {
	httpClient = new(http.Client)
}

type CloudletQueryService struct {
	httpClient *http.Client
}

type WorkloadStatusMessage struct {
	ClientCount int32
}

// NewCloudletQueryService creates *CloudletQueryService
func NewCloudletQueryService() *CloudletQueryService {
	return &CloudletQueryService{
		httpClient: httpClient,
	}
}

// QueryCloudletWorkload queries workload of specific cloudlet
// and returns channel of WorkloadStatusMessage
func (c *CloudletQueryService) QueryCloudletWorkload(ctx context.Context, cloudlet *CloudletNode) <-chan WorkloadStatusMessage {
	outChan := make(chan WorkloadStatusMessage)
	go queryCloudletWorkload(ctx, outChan, c, cloudlet)

	return outChan
}

func queryCloudletWorkload(ctx context.Context, outChan chan WorkloadStatusMessage, c *CloudletQueryService, cn *CloudletNode) {
	queryServiceAddress := "http://" + cn.IPAddr + ":6000/info/currentclient"
	parsedURL, err := url.Parse(queryServiceAddress)
	if err != nil {
		log.Println("cannot parse URL")
		return
	}

	if c.httpClient == nil {
		log.Println("why the fuck http client is nil?")
		return
	}

	response, err := c.httpClient.Get(parsedURL.String())
	if err != nil {
		log.Printf("cannot get service from %s, error: %s", queryServiceAddress, err.Error())
		return
	}
	defer response.Body.Close()
	defer func() {
		close(outChan)
		log.Printf("workload monitoring of cloudlet %v stoped", cn.Name)
	}()

	decoder := json.NewDecoder(response.Body)
	var statusMsg WorkloadStatusMessage
	cancel := ctx.Done()

	for {
		select {
		case <-cancel:
			return
		default:
			if err := decoder.Decode(&statusMsg); err != nil {
				log.Println(err)
				return
			}
			outChan <- statusMsg
		}
	}
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

type CloudletQueryService struct {
	httpClient *http.Client
}

type WorkloadStatusMessage struct {
	ClientCount int32
}

func NewCloudletQueryService() *CloudletQueryService {
	return &CloudletQueryService{
		httpClient: new(http.Client),
	}
}

// QueryCloudletWorkload queries workload of specific cloudlet
// and returns channel of WorkloadStatusMessage
func (c *CloudletQueryService) QueryCloudletWorkload(cloudlet *CloudletNode) <-chan WorkloadStatusMessage {
	outChan := make(chan WorkloadStatusMessage)
	go queryCloudletWorkload(outChan, c, cloudlet)

	return outChan
}

func queryCloudletWorkload(outChan chan WorkloadStatusMessage, c *CloudletQueryService, cn *CloudletNode) {
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
	log.Printf("Call pass response")
	if err != nil {
		log.Printf("Cannot get service from %s, error: %s", queryServiceAddress, err.Error())
		return
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	var statusMsg WorkloadStatusMessage
	for {
		if err := decoder.Decode(&statusMsg); err != nil {
			log.Println(err)
			close(outChan)
			break
		}
		log.Println("read msg")
		outChan <- statusMsg
	}
}

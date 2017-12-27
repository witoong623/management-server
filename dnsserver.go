package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/witoong623/management-server/dnscache"
)

type DNSServer struct {
	dnsServer *dns.Server
	manageCtx *manageCtx
}

func (d *DNSServer) parseQuery(clientIP string, m *dns.Msg) {
	for _, q := range m.Question {
		var err error
		var ip string
		cleanedName := q.Name[0 : len(q.Name)-1] // remove the end "."
		qType := "A"

		// check available service and begin handle request to Cloudlet
		requestedService, err := d.manageCtx.ServiceManager.GetService(cleanedName)
		if err == nil {
			type serviceAndNode struct {
				service      Service
				cloudletNode *CloudletNode
			}
			availableNodes := make([]*serviceAndNode, 0, len(d.manageCtx.Cloudlets))
			// find the cloudlets that have requested service
			for _, cloudletNode := range d.manageCtx.Cloudlets {
				for _, serviceInNode := range cloudletNode.AvailableServices {
					if serviceInNode.Name == requestedService.Name {
						availableNodes = append(availableNodes, &serviceAndNode{service: serviceInNode, cloudletNode: cloudletNode})
						break
					}
				}
			}
			// right now, we got all of nodes that serve that service
			var leastwork *serviceAndNode
			var leastworkvalue int32
			if len(availableNodes) == 1 {
				// only 1 node available
				leastwork = availableNodes[0]
			} else {
				// find least work
				leastwork = availableNodes[0]
				leastworkvalue = leastwork.cloudletNode.GetCurrentWorkload()
				for _, nodeNservice := range availableNodes[1:] {
					currentworkvalue := nodeNservice.cloudletNode.GetCurrentWorkload()
					if currentworkvalue < leastworkvalue {
						leastworkvalue = currentworkvalue
						leastwork = nodeNservice
					}
				}
			}
			// we got 1 Cloudlet that have least works
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, leastwork.cloudletNode.IPAddr))
			if err == nil {
				m.Answer = append(m.Answer, rr)
				return
			}
			// I don't know why err isn't nil so let resolve domain name using normal procedure
		}

		if q.Qtype == dns.TypeA {
			ip, err = dnscache.GetDomainIPv4(cleanedName)
		} else if q.Qtype == dns.TypeAAAA {
			ip, err = dnscache.GetDomainIPv6(cleanedName)
			qType = "AAAA"
		}

		if ip != "" && err == nil {
			rr, err := dns.NewRR(fmt.Sprintf("%s %s %s", q.Name, qType, ip))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		} else {
			// Request to a DNS server
			c := new(dns.Client)
			msg := new(dns.Msg)
			msg.SetQuestion(dns.Fqdn(q.Name), q.Qtype)
			msg.RecursionDesired = true

			r, _, err := c.Exchange(msg, net.JoinHostPort(Config.UpstreamDNSServerAddr, "53"))
			if r == nil {
				log.Printf("*** error: %s\n", err.Error())
				return
			}

			if r.Rcode != dns.RcodeSuccess {
				log.Printf(" *** invalid answer name %s after %s query for %s\n", q.Name, qType, q.Name)
				return
			}
			// Parse Answer
			for _, a := range r.Answer {
				ans := strings.Split(a.String(), "\t")
				if len(ans) == 5 && ans[3] == qType {
					// Save on cache
					if q.Qtype == dns.TypeA {
						dnscache.AddDomainIPv4(cleanedName, ans[4], int(Config.DNSCacheTime))
					} else if q.Qtype == dns.TypeAAAA {
						dnscache.AddDomainIPv6(cleanedName, ans[4], int(Config.DNSCacheTime))
					}
				}
			}
			// Set answer for the client
			m.Answer = r.Answer
		}
	}
}

func (d *DNSServer) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	clientIp := w.RemoteAddr().String()
	clientIp = clientIp[0:strings.LastIndex(clientIp, ":")] // remove port

	switch r.Opcode {
	case dns.OpcodeQuery:
		d.parseQuery(clientIp, m)
	}

	w.WriteMsg(m)
}

// ListenAndServe listens to DNS request.
func (d *DNSServer) ListenAndServe() {
	dns.HandleFunc(".", d.handleDnsRequest)
	log.Printf("start domain name server at IP %v\n", Config.DNSServerAddr)

	err := d.dnsServer.ListenAndServe()
	if err != nil {
		log.Println("DNS ListenAndServe return error: ", err.Error())
	}

}

// Shutdown shutdowns DNS server.
func (d *DNSServer) Shutdown() {
	d.dnsServer.Shutdown()
}

// NewDNSServer creates new instance of DNSServer by providing application context information.
func NewDNSServer(appCtx *manageCtx) *DNSServer {
	dnsserver := &DNSServer{
		dnsServer: &dns.Server{Addr: net.JoinHostPort(Config.DNSServerAddr, strconv.FormatInt(int64(Config.DNSPortNumber), 10)), Net: "udp"},
		manageCtx: appCtx,
	}

	return dnsserver
}

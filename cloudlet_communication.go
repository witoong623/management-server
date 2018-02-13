package main

import (
	"log"
	"net/http"
)

// HandleCloudletRegisterCommand handles monitor request.
func HandleCloudletRegisterCommand(c *manageCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("accept only POST method"))
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		formData := r.Form
		cName := formData.Get("cloudlet-name")
		cIP := formData.Get("cloudlet-ip")
		cDomain := formData.Get("cloudlet-domain")
		log.Printf("got register request from %v, IP %v and domain %v\n", cName, cIP, cDomain)

		cNode, err := NewCloudletNode(cName, cIP, cDomain)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("cloudlet %v has been registered\n", cName)

		c.Cloudlets[cName] = cNode
	})
}

// HandleServiceRegisterCommand handles service registration from cloudlet.
// Cloudlet will send service name and domain name associated with service to management server
// and management server will check if specified "service" has been added to management server
// if it haven't been added, new service instance will be created and added to management server.
// Regardless of service creation, specified service will be associated with cloudlet.
// However, cloudlet should register itself to management server.  If request come from cloudlet
// that haven't been registered, reject request.
func HandleServiceRegisterCommand(c *manageCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("accept only POST method"))
			return
		}
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		formData := r.Form
		cName := formData.Get("cloudlet-name")
		if cName == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("cloudlet-name is empty"))
			return
		}
		cloudlet, ok := c.Cloudlets[cName]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("cloudlet must register itself with management server before register service"))
			return
		}

		sDomain := formData.Get("service-domain")
		sName := formData.Get("service-name")
		if sDomain == "" || sName == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("service-domain or service-name is empty"))
			return
		}
		service, err := c.ServiceManager.GetService(sDomain)
		if err != nil {
			service = &Service{Name: sName, Domain: sDomain}
			c.ServiceManager.AddService(sDomain, service)
			log.Printf("service %v has been added\n", sName)
		}

		cloudlet.AvailableServices = append(cloudlet.AvailableServices, service)
		log.Printf("service %v has been associated with cloudlet %v", sName, cName)
	})
}

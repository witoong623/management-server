package main

import "net/http"

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
		}

		formData := r.Form
		cName := formData.Get("name")
		cIP := formData.Get("ip")
		cDomain := formData.Get("domain")

		cNode, err := NewCloudletNode(cName, cIP, cDomain)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		c.Cloudlets[cName] = cNode
	})
}

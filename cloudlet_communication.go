package main

import "net/http"

// HandleCloudletRegisterCommand handles monitor request.
func HandleCloudletRegisterCommand(c *edgeProxyCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
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

		_, err := NewCloudletNode(cName, cIP, cDomain)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}

package proxy

import (
	"comet/internal/config"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var URL string

	for _, backend := range config.Backends {
		if backend.Address == "" {
			continue
		}

		if matchRoute(backend.RouteFilter, r.URL.RequestURI()) {
			var route string
			var err error

			if backend.StripPrefix {
				route, err = removeFilterPrefix(backend.RouteFilter, r.URL.RequestURI())
				if err != nil {
					fmt.Println("ERROR: invalid route filter")
					return
				}
			} else {
				route = r.URL.RequestURI()
			}

			URL, err = url.JoinPath(backend.Address, route)
			if err != nil {
				fmt.Println("ERROR: invalid route filter")
				return
			}
		}
	}

	// no URL matched the request
	if URL == "" {
		http.Error(w, "NOT FOUND", http.StatusNotFound)
		return
	}

	resp, err := http.Get(URL)
	if err != nil {
		http.Error(w, "Unable to access backend server", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, value := range resp.Header {
		for _, val := range value {
			w.Header().Add(key, val)
		}
	}

	w.Header().Add("X-Forwarded-For", r.RemoteAddr)
	w.Header().Add("X-Forwarded-Host", r.URL.RawPath)

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Unable to send response data", http.StatusInternalServerError)
		return
	}

}

func StartProxy() error {
	http.HandleFunc("/", handleRequest)

	return http.ListenAndServe(viper.GetString("proxy_address"), nil)
}

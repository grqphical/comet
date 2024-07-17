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
	url, err := url.JoinPath(config.Backends[0].Address, r.URL.RequestURI())
	if err != nil {
		fmt.Println("ERROR: invalid URL for backend")
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

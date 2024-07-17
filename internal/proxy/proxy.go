package proxy

import (
	"comet/internal/config"
	"comet/internal/logging"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Proxy struct {
	backendStatus map[string]bool
	mu            sync.Mutex
}

func NewProxy() Proxy {
	var serversStatus = make(map[string]bool)

	for _, backend := range config.Backends {
		url, err := url.JoinPath(backend.Address, backend.HealthEndpoint)
		if err != nil {
			logging.LogCritical("invalid address or health endpoint")
		}
		_, err = http.Get(url)

		if err == nil {
			serversStatus[backend.Address] = true
		} else {
			serversStatus[backend.Address] = false
			logging.Logger.Warn("server offline", "address", backend.Address)
		}

	}

	return Proxy{
		backendStatus: serversStatus,
	}
}

func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var URL string

	for _, backend := range config.Backends {
		if backend.Address == "" {
			logging.Logger.Warn("backend has no configured address")
			continue
		}

		p.mu.Lock()
		if !p.backendStatus[backend.Address] {
			p.mu.Unlock()
			http.Error(w, "backend not avaliable", http.StatusServiceUnavailable)
			return
		}
		p.mu.Unlock()

		if matchRoute(backend.RouteFilter, r.URL.RequestURI()) {
			var route string
			var err error

			if backend.StripFilter {
				route, err = removeFilterPrefix(backend.RouteFilter, r.URL.RequestURI())
				if err != nil {
					logging.LogCritical("invalid URL filter")
				}
			} else {
				route = r.URL.RequestURI()
			}

			URL, err = url.JoinPath(backend.Address, route)
			if err != nil {
				logging.LogCritical("invalid URL filter")
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
	responseTime := time.Since(startTime)

	if viper.GetBool("log_requests") {
		logging.Logger.Info("", "method", r.Method, "status", resp.StatusCode, "route", r.RequestURI, "ip", r.RemoteAddr, "responseTime", responseTime.Microseconds())
	}

}

func (p *Proxy) checkHealth() {
	p.mu.Lock()
	for _, backend := range config.Backends {
		url, err := url.JoinPath(backend.Address, backend.HealthEndpoint)
		if err != nil {
			logging.LogCritical("invalid address or health endpoint")
		}
		_, err = http.Get(url)

		if err == nil {
			p.backendStatus[backend.Address] = true
		} else {
			p.backendStatus[backend.Address] = false
			logging.Logger.Warn("server offline", "address", backend.Address)
		}

	}
	p.mu.Unlock()
}

func (p *Proxy) StartProxy() error {
	http.HandleFunc("/", p.handleRequest)

	go func() {
		ticker := time.NewTicker(time.Second * 15)
		for range ticker.C {
			p.checkHealth()
		}
	}()

	return http.ListenAndServe(viper.GetString("proxy_address"), nil)
}

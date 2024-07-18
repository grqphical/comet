package server

import (
	"comet/internal/config"
	"comet/internal/logging"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type Proxy struct {
	serverStatuses map[string]bool
	backend        *config.Backend
	mu             sync.Mutex
}

func newProxy(backend *config.Backend) *Proxy {
	var serversStatus = make(map[string]bool)

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

	return &Proxy{
		serverStatuses: serversStatus,
		backend:        backend,
	}
}

func (p *Proxy) HandleRequest(w http.ResponseWriter, r *http.Request) int {
	var URL string

	if p.backend.Address == "" {
		logging.Logger.Warn("backend has no configured address")
	}

	p.mu.Lock()
	if !p.serverStatuses[p.backend.Address] {
		p.mu.Unlock()
		http.Error(w, "backend not avaliable", http.StatusServiceUnavailable)
		return 503
	}
	p.mu.Unlock()

	if matchRoute(p.backend.RouteFilter, r.URL.RequestURI()) {
		var route string
		var err error

		if p.backend.StripFilter {
			route, err = removeFilterPrefix(p.backend.RouteFilter, r.URL.RequestURI())
			if err != nil {
				logging.LogCritical("invalid URL filter")
			}
		} else {
			route = r.URL.RequestURI()
		}

		for _, hiddenRoute := range p.backend.HiddenRoutes {
			if matchRoute(hiddenRoute, route) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return 403
			}
		}

		URL, err = url.JoinPath(p.backend.Address, route)
		if err != nil {
			logging.LogCritical("invalid URL filter")
		}
	}

	// no URL matched the request
	if URL == "" {
		http.Error(w, "NOT FOUND", http.StatusNotFound)
		return 404
	}

	resp, err := http.Get(URL)
	if err != nil {
		http.Error(w, "Unable to access backend server", http.StatusInternalServerError)
		return 500
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
		return 500
	}

	return resp.StatusCode
}

func (p *Proxy) CheckHealth() {
	p.mu.Lock()

	url, err := url.JoinPath(p.backend.Address, p.backend.HealthEndpoint)
	if err != nil {
		logging.LogCritical("invalid backend address/health endpoint")
	}
	_, err = http.Get(url)

	if err == nil {
		p.serverStatuses[p.backend.Address] = true
	} else {
		p.serverStatuses[p.backend.Address] = false
		logging.Logger.Warn("server offline", "address", p.backend.Address)
	}

	p.mu.Unlock()
}

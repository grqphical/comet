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

	if backend.CheckHealth {
		_, err = http.Get(url)

		if err == nil {
			serversStatus[backend.Address] = true
		} else {
			serversStatus[backend.Address] = false
			logging.Logger.Warn("server offline", "address", backend.Address)
		}
	} else {
		serversStatus[backend.Address] = true
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

		sendError(503, w)
		return 503
	}
	p.mu.Unlock()

	var route string
	var err error

	if p.backend.StripFilter {
		route, _ = removeFilterPrefix(p.backend.RouteFilter, r.URL.RequestURI())
	} else {
		route = r.URL.RequestURI()
	}

	for _, hiddenRoute := range p.backend.HiddenRoutes {
		if matchRoute(hiddenRoute, route) {
			sendError(403, w)
			return 403
		}
	}

	URL = p.backend.Address + route

	// no URL matched the request
	if URL == "" {
		sendError(404, w)
		return 404
	}

	request, err := http.NewRequest(r.Method, URL, nil)
	if err != nil {
		sendError(500, w)
	}

	request.Header.Add("X-Forwarded-For", r.RemoteAddr)
	request.Header.Add("X-Forwarded-Host", r.Host)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		sendError(500, w)
		return 500
	}
	defer resp.Body.Close()

	for key, value := range resp.Header {
		for _, val := range value {
			w.Header().Add(key, val)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		sendError(500, w)
		return 500
	}

	return resp.StatusCode
}

func (p *Proxy) CheckHealth() {
	if !p.backend.CheckHealth {
		return
	}

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

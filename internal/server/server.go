package server

import (
	"comet/internal/config"
	"comet/internal/logging"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type Backend interface {
	HandleRequest(http.ResponseWriter, *http.Request)
	RouteFilter() string
	CheckHealth()
}

type Server struct {
	Backends map[string]Backend
}

func NewServer() Server {
	servers := make(map[string]Backend)

	for _, backend := range config.Backends {
		switch backend.Type {
		case "proxy":
			servers[backend.RouteFilter] = newProxy(&backend)
		default:
			logging.Logger.Error("unknown type defined in configuration. Valid types are proxy and staticfs")
		}
	}

	return Server{
		Backends: servers,
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	for _, backend := range s.Backends {
		if matchRoute(backend.RouteFilter(), r.URL.RequestURI()) {
			backend.HandleRequest(w, r)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) StartServer() error {
	http.HandleFunc("/", s.handleRequest)

	go func() {
		ticker := time.NewTicker(time.Second * 15)
		for range ticker.C {
			for _, backend := range s.Backends {
				backend.CheckHealth()
			}
		}
	}()

	return http.ListenAndServe(viper.GetString("proxy_address"), nil)
}

package server

import (
	"comet/internal/config"
	"comet/internal/logging"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type Handler interface {
	HandleRequest(http.ResponseWriter, *http.Request)
	CheckHealth()
}

type Server struct {
	Handlers map[string]Handler
}

func NewServer() Server {
	servers := make(map[string]Handler)

	for _, backend := range config.Backends {
		switch backend.Type {
		case "proxy":
			servers[backend.RouteFilter] = newProxy(&backend)
		case "staticfs":
			servers[backend.RouteFilter] = newStaticFS(&backend)
		default:
			logging.Logger.Warn("unknown type defined in configuration, ignoring backend")
		}
	}

	return Server{
		Handlers: servers,
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	for filter, backend := range s.Handlers {
		if matchRoute(filter, r.URL.RequestURI()) {
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
			for _, backend := range s.Handlers {
				backend.CheckHealth()
			}
		}
	}()

	logging.Logger.Info("starting server", "address", viper.GetString("proxy_address"))

	return http.ListenAndServe(viper.GetString("proxy_address"), nil)
}

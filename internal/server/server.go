package server

import (
	"comet/internal/config"
	"comet/internal/logging"
	"net"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type Handler interface {
	HandleRequest(http.ResponseWriter, *http.Request) int
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
	for _, ip := range viper.GetStringSlice("ip_filter.blacklist") {
		netIP := net.ParseIP(ip)
		incomingIP := net.ParseIP(r.RemoteAddr)

		if net.IP.Equal(netIP, incomingIP) {
			sendError(http.StatusForbidden, w)
			return
		}
	}

	for filter, backend := range s.Handlers {
		if matchRoute(filter, r.URL.RequestURI()) {
			startTime := time.Now()
			status := backend.HandleRequest(w, r)
			if viper.GetBool("log_requests") {
				responseTime := time.Since(startTime)
				logging.Logger.Info("proxy", "method", r.Method, "status", status, "route", r.RequestURI, "ip", r.RemoteAddr, "responseTime", responseTime.Microseconds())
			}
			return
		}
	}

	sendError(http.StatusNotFound, w)
}

func (s *Server) StartServer() error {
	http.HandleFunc("/", s.handleRequest)

	go func() {
		duration := viper.GetInt("health_check_interval")
		if duration == 0 {
			return
		}

		ticker := time.NewTicker(time.Second * time.Duration(duration))
		for range ticker.C {
			for _, backend := range s.Handlers {
				backend.CheckHealth()
			}
		}
	}()

	logging.Logger.Info("starting server", "address", viper.GetString("proxy_address"))

	return http.ListenAndServe(viper.GetString("proxy_address"), nil)
}

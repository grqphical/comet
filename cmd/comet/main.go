package main

import (
	"comet/internal/config"
	"comet/internal/logging"
	"comet/internal/server"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		logging.LogCritical(err.Error())
	}

	p := server.NewServer()

	err = p.StartServer()
	if err != nil {
		logging.LogCritical(err.Error())
	}
}

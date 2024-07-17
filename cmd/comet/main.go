package main

import (
	"comet/internal/config"
	"comet/internal/logging"
	"comet/internal/proxy"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		logging.LogCritical(err.Error())
	}

	p := proxy.NewProxy()

	err = p.StartProxy()
	if err != nil {
		logging.LogCritical(err.Error())
	}
}

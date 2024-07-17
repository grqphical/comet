package main

import (
	"comet/internal/config"
	"comet/internal/proxy"
	"fmt"
	"os"
)

func LogError(err error) {
	fmt.Printf("ERROR: %s\n", err)
	os.Exit(1)
}

func main() {
	err := config.ReadConfig()
	if err != nil {
		LogError(err)
	}

	err = proxy.StartProxy()
	if err != nil {
		LogError(err)
	}
}

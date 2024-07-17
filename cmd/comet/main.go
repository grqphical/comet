package main

import (
	"comet/internal/config"
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
}

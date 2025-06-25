package main

import (
	"fmt"
	"github.com/crafty-ezhik/amocrmproxy/config"
	"github.com/crafty-ezhik/amocrmproxy/logger"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.Debug)
	log.Info("Starting proxy server")
	fmt.Printf("%#v\n", cfg)
}

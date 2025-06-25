package main

import (
	"fmt"
	"github.com/crafty-ezhik/amocrmproxy/config"
	"github.com/crafty-ezhik/amocrmproxy/logger"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.Debug)
	log.Info("Starting proxy server")
	fmt.Printf("%#v\n", cfg)

	router := chi.NewRouter()
	
}

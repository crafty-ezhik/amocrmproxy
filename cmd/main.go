package main

import (
	"fmt"
	"github.com/crafty-ezhik/amocrmproxy/config"
	"github.com/crafty-ezhik/amocrmproxy/handlers"
	"github.com/crafty-ezhik/amocrmproxy/logger"
	"github.com/crafty-ezhik/amocrmproxy/routes"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.Debug)
	log.Info("Starting proxy server")
	fmt.Printf("%#v\n", cfg)

	appHandlers := handlers.NewAppHandlers(log, cfg)

	router := chi.NewRouter()
	routes.InitMiddleware(router)
	routes.InitRoutes(router, appHandlers)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.ServerPort),
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Error("Error starting server")
	}

}

package main

import (
	"fmt"
	"github.com/crafty-ezhik/amocrmproxy/config"
	"github.com/crafty-ezhik/amocrmproxy/handlers"
	"github.com/crafty-ezhik/amocrmproxy/logger"
	"github.com/crafty-ezhik/amocrmproxy/routes"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.Debug)

	// Инициализация хендлера
	appHandlers := handlers.NewAppHandlers(log, cfg)

	// Инициализация роутера, Middleware и маршрутов
	router := chi.NewRouter()
	routes.InitMiddleware(router, cfg.Server.RequestTimeout)
	routes.InitRoutes(router, appHandlers)

	// Конфигурирование сервера
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.ServerPort),
		Handler: router,
	}

	// Старт сервера
	err := server.ListenAndServe()
	if err != nil {
		log.Error("Error starting server")
	}
	log.Info("Starting proxy server on port: " + strconv.Itoa(cfg.Server.ServerPort))

}

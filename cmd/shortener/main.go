package main

import (
	"fmt"
	"net/http"

	"github.com/Tokebay/yandex/config"
	"github.com/Tokebay/yandex/internal/app"

	logger "github.com/Tokebay/yandex/internal/logger"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {
	logger.Initialize("info")

	cfg := config.NewConfig()
	storage := app.NewMapStorage()
	shortener := app.NewURLShortener(cfg, storage)

	r := chi.NewRouter()
	r.Use(logger.LoggerMiddleware)

	r.Post("/", shortener.ShortenURLHandler)
	r.Get("/{id}", shortener.RedirectURLHandler)

	addr := fmt.Sprintf("%s", cfg.ServerAddress)
	logger.Log.Info("Server is starting", zap.String("address", addr))

	err := http.ListenAndServe(addr, r)
	if err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
	}
}

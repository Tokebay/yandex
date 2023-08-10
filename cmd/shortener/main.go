package main

import (
	"fmt"
	"net/http"

	"github.com/Tokebay/yandex/config"
	"github.com/Tokebay/yandex/internal/app"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()
	storage := app.NewMapStorage()
	shortener := app.NewURLShortener(cfg, storage)

	r := chi.NewRouter()
	r.Post("/", shortener.ShortenURLHandler)
	r.Get("/{id}", shortener.RedirectURLHandler)

	fmt.Printf("baseUrl %s\n", cfg.BaseURL)

	serverAddress := cfg.ServerAddress

	fmt.Printf("Server is running on http://%s\n", serverAddress)
	http.ListenAndServe(serverAddress, r)
}

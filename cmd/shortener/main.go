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
	if err := run(); err != nil {
		fmt.Println("Error", err)
	}
}

func run() error {
	//Инициализируется логгер
	logger.Initialize("info")

	cfg := config.NewConfig()
	storage := app.NewMapStorage()
	var fileStorage *app.Producer
	var shortener *app.URLShortener
	var err error

	fmt.Printf("FileStoragePath: %s\n", cfg.FileStoragePath)
	if cfg.FileStoragePath != "" {
		fileStorage, err = app.NewProducer(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Error("Error in NewProducer", zap.Error(err))
			return err
		}
		defer fileStorage.Close()

		urlDataSlice, err := app.LoadURLsFromFile(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Error("Error in LoadURLsFromFile", zap.Error(err))
			return err
		}
		shortener = app.NewURLShortener(cfg, storage, fileStorage)
		shortener.URLDataSlice = urlDataSlice
	} else {
		shortener = app.NewURLShortener(cfg, storage, fileStorage)
	}
	// маршрутизатор (chi.Router), который будет использоваться для обработки HTTP-запросов.
	r := chi.NewRouter()
	// промежуточное ПО (middleware) для логирования. перед каждым запросом будет выполнена функция logger.LoggerMiddleware
	r.Use(logger.LoggerMiddleware)
	r.Use(logger.RecoveryMiddleware)
	// middleware проверяет поддержку сжатия gzip
	r.Use(app.GzipMiddleware)

	r.Post("/", shortener.ShortenURLHandler)
	r.Get("/{id}", shortener.RedirectURLHandler)
	r.Post("/api/shorten", shortener.APIShortenerURL)

	addr := cfg.ServerAddress
	logger.Log.Info("Server is starting", zap.String("address", addr))

	// Запускается HTTP-сервер, который начинает прослушивание указанного адреса addr и использует маршрутизатор r для обработки запросов.
	err = http.ListenAndServe(addr, r)
	if err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
		return err
	}

	return nil
}

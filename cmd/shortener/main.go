package main

import (
	"fmt"
	"net/http"

	"github.com/Tokebay/yandex/config"

	"github.com/Tokebay/yandex/internal/app/handlers"
	"github.com/Tokebay/yandex/internal/app/storage"
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

	// конфиг для запуска
	cfg := config.NewConfig()

	storage := storage.NewMapStorage()
	var fileStorage *handlers.Producer
	var shortener *handlers.URLShortener

	fmt.Printf("FileStoragePath: %s; DSN: %s\n", cfg.FileStoragePath, cfg.DataBaseConnString)

	// если флаг -d пустой пропускаем и сохраняем все в map и файл
	if cfg.DataBaseConnString == "" {

		if cfg.FileStoragePath != "" {
			fileStorage, err := handlers.NewProducer(cfg.FileStoragePath)
			if err != nil {
				logger.Log.Error("Error in NewProducer", zap.Error(err))
				return err
			}
			defer fileStorage.Close()

			urlDataSlice, err := fileStorage.LoadInitialData()
			if err != nil {
				logger.Log.Error("Error loading data from file", zap.Error(err))
				return err
			}

			for _, urlData := range urlDataSlice {
				// fmt.Printf("urlData.ShortURL %s;  urlData.OriginalUR %s \n", urlData.ShortURL, urlData.OriginalURL)
				err := storage.SaveMapURL(urlData.ShortURL, urlData.OriginalURL)
				if err != nil {
					logger.Log.Error("Error saving URL to storage", zap.Error(err))
					return err
				}
			}
			shortener = handlers.NewURLShortener(cfg, storage, fileStorage)
		}
	} else {
		// если флаг -f пустой
		shortener = handlers.NewURLShortener(cfg, storage, fileStorage)
	}
	// маршрутизатор (chi.Router), который будет использоваться для обработки HTTP-запросов.
	r := chi.NewRouter()
	// промежуточное ПО (middleware) для логирования. перед каждым запросом будет выполнена функция logger.LoggerMiddleware
	r.Use(logger.LoggerMiddleware)
	r.Use(logger.RecoveryMiddleware)
	// middleware проверяет поддержку сжатия gzip
	r.Use(handlers.GzipMiddleware)

	r.Post("/", shortener.ShortenURLHandler)
	r.Get("/{id}", shortener.RedirectURLHandler)
	r.Post("/api/shorten", shortener.APIShortenerURL)
	r.Get("/ping", shortener.CheckDBConnect)

	addr := cfg.ServerAddress
	logger.Log.Info("Server is starting", zap.String("address", addr))

	// Запускается HTTP-сервер, который начинает прослушивание указанного адреса addr и использует маршрутизатор r для обработки запросов.
	err := http.ListenAndServe(addr, r)
	if err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
		return err
	}

	return nil
}

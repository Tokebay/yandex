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

	mapStorage := storage.NewMapStorage()
	// var storage storage.URLStorage
	var fileStorage *handlers.Producer
	var shortener *handlers.URLShortener
	var err error
	fmt.Printf("FileStoragePath: %s; DSN: %s \n", cfg.FileStoragePath, cfg.DSN)

	if cfg.DSN != "" {
		fmt.Println("connect to DB")

		// Инициализировать и использовать PostgreSQL хранилище
		dbStorage, err := storage.NewPostgreSQLStorage(cfg.DSN)
		if err != nil {
			logger.Log.Error("Error in NewPostgreSQLStorage", zap.Error(err))
			return err
		}

		err = dbStorage.CreateTable()
		if err != nil {
			logger.Log.Error("Error creating table in PostgreSQL", zap.Error(err))
			return err
		}
		shortener = handlers.NewURLShortener(cfg, dbStorage, nil)

	} else {
		fileStorage, err = handlers.NewProducer(cfg.FileStoragePath)
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
			err := mapStorage.SaveURL(urlData.ShortURL, urlData.OriginalURL)
			if err != nil {
				logger.Log.Error("Error saving URL to storage", zap.Error(err))
				return err
			}
		}
		shortener = handlers.NewURLShortener(cfg, mapStorage, fileStorage)
	}

	r := createRouter(shortener, cfg)
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

func createRouter(shortener *handlers.URLShortener, cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	// Промежуточное ПО (middleware) для логирования. перед каждым запросом будет выполнена функция logger.LoggerMiddleware
	r.Use(logger.LoggerMiddleware)
	r.Use(logger.RecoveryMiddleware)
	// middleware проверяет поддержку сжатия gzip
	r.Use(handlers.GzipMiddleware)

	r.Post("/", shortener.ShortenURLHandler)
	r.Get("/{id}", shortener.RedirectURLHandler)
	r.Post("/api/shorten", shortener.APIShortenerURL)
	r.Get("/ping", shortener.CheckDBConnect)

	return r
}

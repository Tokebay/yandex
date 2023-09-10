package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tokebay/yandex/config"

	"github.com/Tokebay/yandex/internal/app/handlers"
	"github.com/Tokebay/yandex/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestURLShortener_shortenURLHandler(t *testing.T) {

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
	}
	storage := *storage.NewMapStorage()
	fileStorage, err := handlers.NewProducer(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileStorage.Close()
	shortener := handlers.NewURLShortener(cfg, &storage, fileStorage)

	// Устанавливаем функцию генерации идентификатора для тестов
	shortener.SetGenerateIDFunc(func() string {
		return "EwHXdJfB"
	})

	type want struct {
		contentType string
		statusCode  int
		shortURL    string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "POST_ShortenUrl",
			request: "https://practicum.yandex.ru/",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
				shortURL:    "http://localhost:8080/EwHXdJfB",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := strings.NewReader(tt.request)
			request := httptest.NewRequest(http.MethodPost, "/", requestBody)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			shortener.ShortenURLHandler(w, request)

			res := w.Result()
			// проверяем статус код
			assert.Equal(t, res.StatusCode, tt.want.statusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			bodyContent, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.shortURL, string(bodyContent))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func TestApiShortenerURL(t *testing.T) {
	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
	}
	storage := *storage.NewMapStorage()
	fileStorage, err := handlers.NewProducer(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileStorage.Close()
	shortener := handlers.NewURLShortener(cfg, &storage, fileStorage)

	// Устанавливаем функцию генерации идентификатора для тестов
	shortener.SetGenerateIDFunc(func() string {
		return "EwHXdJfB"
	})

	type want struct {
		contentType  string
		statusCode   int
		expectedBody string
	}
	tests := []struct {
		name        string
		contentType string
		requestBody []byte
		want        want
	}{
		{
			name:        "JSON_ApiShortenerURL",
			contentType: "application/json",
			requestBody: []byte(`{ "url": "https://practicum.yandex.ru"}`),

			want: want{
				contentType:  "application/json",
				statusCode:   201,
				expectedBody: `{"result":"http://localhost:8080/EwHXdJfB"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(tt.requestBody))
			httptest.NewRecorder().Header().Set("Content-Type", "application/json")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			shortener.APIShortenerURL(w, request)

			res := w.Result()
			// проверяем статус код
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			bodyContent := w.Body.String()
			assert.Equal(t, tt.want.expectedBody, bodyContent)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestRedirectURLHandler_redirectURLHandler(t *testing.T) {
	storage := storage.NewMapStorage()
	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
	}

	fileStorage, err := handlers.NewProducer(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	shortener := handlers.NewURLShortener(
		cfg,
		storage,
		fileStorage,
	)

	storage.SaveURL("BpLnfgSfEr", "https://mail.ru/")

	type want struct {
		statusCode  int
		originalURL string
	}
	tests := []struct {
		name     string
		shortURL string
		want     want
	}{
		{
			name:     "RedirectURL",
			shortURL: "http://localhost:8080/BpLnfgSfEr",
			want: want{
				statusCode:  307,
				originalURL: "https://mail.ru/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.shortURL, nil)
			w := httptest.NewRecorder()
			shortener.RedirectURLHandler(w, request)

			res := w.Result()
			defer res.Body.Close()

			// Проверяем статус-код
			assert.Equal(t, tt.want.statusCode, 307)
			// Получаем и проверяем заголовок Location
			assert.Equal(t, tt.want.originalURL, res.Header.Get("Location"))
		})
	}
}

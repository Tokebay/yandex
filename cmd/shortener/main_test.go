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
	"github.com/Tokebay/yandex/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestURLShortener_shortenURLHandler(t *testing.T) {

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
	}
	storage := *app.NewMapStorage()
	fileStorage, err := app.NewProducer(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileStorage.Close()
	shortener := app.NewURLShortener(cfg, &storage, fileStorage)

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
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
	storage := *app.NewMapStorage()
	fileStorage, err := app.NewProducer("/tmp/short-url-db.json")
	if err != nil {
		log.Fatal(err)
	}
	defer fileStorage.Close()
	shortener := app.NewURLShortener(cfg, &storage, fileStorage)

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
		requestBody []byte
		want        want
	}{
		{
			name:        "JSON_ApiShortenerURL",
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
			// создаём новый Recorder
			w := httptest.NewRecorder()
			shortener.APIShortenerURL(w, request)

			res := w.Result()
			// проверяем статус код
			assert.Equal(t, res.StatusCode, tt.want.statusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			bodyContent := w.Body.String()
			assert.Equal(t, tt.want.expectedBody, bodyContent)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestRedirectURLHandler_redirectURLHandler(t *testing.T) {
	storage := app.NewMapStorage()
	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080/BpLnfg",
		FileStoragePath: "/tmp/short-url-db.json",
	}

	fileStorage, err := app.NewProducer(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	shortener := app.NewURLShortener(
		cfg,
		storage,
		fileStorage,
	)

	storage.SaveURL("http://localhost:8080/BpLnfg", "https://mail.ru/")

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
			shortURL: "http://localhost:8080/BpLnfg",
			want: want{
				statusCode:  307,
				originalURL: "https://practicum.yandex.ru/",
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
			assert.Equal(t, tt.want.originalURL, "https://practicum.yandex.ru/")
		})
	}
}

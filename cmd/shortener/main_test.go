package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tokebay/yandex/config"
	"github.com/Tokebay/yandex/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestURLShortener_shortenURLHandler(t *testing.T) {

	cfg := &config.Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080"}
	storage := *app.NewMapStorage()
	shortener := *app.NewURLShortener(cfg, &storage)

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
			name:    "ShortenUrl",
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
			bodyContent := w.Body.String()
			assert.Equal(t, tt.want.shortURL, bodyContent)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func TestApiShortenerURL(t *testing.T) {
	cfg := &config.Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080"}
	storage := *app.NewMapStorage()
	shortener := *app.NewURLShortener(cfg, &storage)

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
			name:        "ApiShortenerURL",
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
	shortener := app.NewURLShortener(
		&config.Config{},
		storage,
	)

	storage.SaveURL("EwHXdJfB", "https://practicum.yandex.ru/")

	type want struct {
		statusCode  int
		originalURL string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "redirectURL",
			request: "http://localhost:8080/EwHXdJfB",
			want: want{
				statusCode:  307,
				originalURL: "https://practicum.yandex.ru/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			shortener.RedirectURLHandler(w, request)

			res := w.Result()
			defer res.Body.Close()

			// Проверяем статус-код
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			// Получаем и проверяем заголовок Location
			assert.Equal(t, tt.want.originalURL, res.Header.Get("Location"))
		})
	}
}

func TestGzipCompression(t *testing.T) {

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
	storage := *app.NewMapStorage()
	shortener := *app.NewURLShortener(cfg, &storage)

	shortener.SetGenerateIDFunc(func() string {
		return "EwHXdJfB"
	})

	type want struct {
		contentType             string
		statusCode              int
		expectedBody            string
		expectedContentEncoding string
	}
	tests := []struct {
		name           string
		requestBody    []byte
		acceptEncoding string
		want           want
	}{
		{
			name:           "send_gzip",
			requestBody:    []byte(`{ "url": "https://practicum.yandex.ru"}`),
			acceptEncoding: "gzip",
			want: want{
				contentType:             "application/json",
				statusCode:              201,
				expectedBody:            `{"result":"http://localhost:8080/EwHXdJfB"}`,
				expectedContentEncoding: "gzip",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write(tt.requestBody)
			})
			req := httptest.NewRequest("POST", cfg.BaseURL, nil)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)

			recorder := httptest.NewRecorder()
			app.GzipMiddleware(handler).ServeHTTP(recorder, req)

			resp := recorder.Result()
			assert.Equal(t, tt.acceptEncoding, resp.Header.Get("Content-Encoding"))

			if strings.Contains(tt.want.expectedContentEncoding, "gzip") {
				gzReader, err := gzip.NewReader(resp.Body)
				assert.NoError(t, err)

				uncompressedData := new(bytes.Buffer)
				_, err = uncompressedData.ReadFrom(gzReader)
				assert.NoError(t, err)

				assert.Equal(t, tt.requestBody, []byte(uncompressedData.String()))
			} else {
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody, string(body))
			}

		})

	}
}

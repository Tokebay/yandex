package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// POST / HTTP/1.1
// Host: localhost:8080
// Content-Type: text/plain
// https://practicum.yandex.ru/

// HTTP/1.1 201 Created
// Content-Type: text/plain
// Content-Length: 30
// http://localhost:8080/EwHXdJfB

func TestURLShortener_shortenURLHandler(t *testing.T) {
	shortener := &URLShortener{
		mapping: make(map[string]string),
	}

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
			shortener.shortenURLHandler(w, request)

			res := w.Result()
			//check status code
			assert.Equal(t, res.StatusCode, tt.want.statusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			bodyContent := w.Body.String()
			assert.Equal(t, tt.want.shortURL, bodyContent)

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

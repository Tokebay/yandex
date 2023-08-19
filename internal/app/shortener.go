package app

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/Tokebay/yandex/config"
)

const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLShortener struct {
	generateIDFunc func() string
	config         *config.Config
	storage        URLStorage
}

// Метод для установки функции генерации идентификатора
func (us *URLShortener) SetGenerateIDFunc(fn func() string) {
	us.generateIDFunc = fn
}

func NewURLShortener(cfg *config.Config, storage URLStorage) *URLShortener {
	return &URLShortener{
		config:  cfg,
		storage: storage,
	}
}

func (us *URLShortener) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := us.config
	url, err := io.ReadAll(r.Body)
	defer r.Body.Close() // закрыл тело запроса

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Генерируем случайный идентификатор для сокращения URL
	id := us.GenerateID()
	shortenedURL := cfg.BaseURL + "/" + id

	err = us.storage.SaveURL(id, string(url))
	if err != nil {
		http.Error(w, "Error saving URL", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Original URL: %s\n", url)
	fmt.Printf("Shortened URL: %s\n", shortenedURL)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(shortenedURL)))
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortenedURL))
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return // возврат из функции после обработки ошибки
	}
}

func (us *URLShortener) RedirectURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	// fmt.Printf("Received id: %s\n", id)

	originalURL, err := us.storage.GetURL(id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	// Выполняем перенаправление на оригинальный URL
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (us *URLShortener) GenerateID() string {
	if us.generateIDFunc != nil {
		return us.generateIDFunc()
	}

	base := len(base62Alphabet)
	var idBuilder strings.Builder
	// Генерируем случайный идентификатор из 6 символов
	for i := 0; i < 6; i++ {
		index := rand.Intn(base)
		idBuilder.WriteByte(base62Alphabet[index])
	}

	return idBuilder.String()
}

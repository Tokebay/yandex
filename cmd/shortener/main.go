package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLShortener struct {
	mapping        map[string]string
	generateIDFunc func() string
}

// Метод для установки функции генерации идентификатора
func (us *URLShortener) SetGenerateIDFunc(fn func() string) {
	us.generateIDFunc = fn
}

func main() {
	shortener := &URLShortener{
		mapping: make(map[string]string),
	}

	r := chi.NewRouter()
	r.Post("/", shortener.shortenURLHandler)
	r.Get("/{id}", shortener.redirectURLHandler)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func (us *URLShortener) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Генерируем случайный идентификатор для сокращения URL
	id := us.generateID()
	shortenedURL := "http://localhost:8080/" + id

	us.mapping[id] = string(url)

	fmt.Printf("Original URL: %s\n", url)
	fmt.Printf("Shortened URL: %s\n", shortenedURL)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(shortenedURL)))
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortenedURL))
	if err != nil {
		fmt.Printf("Возникла ошибка %s", err)
	}
}

func (us *URLShortener) redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, ok := us.mapping[id]
	if !ok {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	// Выполняем перенаправление на оригинальный URL
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (us *URLShortener) generateID() string {
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

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/Tokebay/yandex/config"

	"github.com/Tokebay/yandex/internal/app/storage"
	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	"go.uber.org/zap"
)

const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLShortener struct {
	generateIDFunc func() string
	config         *config.Config
	Storage        storage.URLStorage
	fileStorage    *Producer
	uuidCounter    int // счетчик UUID
	uuidMu         sync.Mutex
	URLDataSlice   []URLData
}

type URLData struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// func LoadURLsFromFile(filePath string) ([]URLData, error) {

// 	fmt.Printf("filePath: %s\n", filePath)

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		logger.Log.Info("Error os.Open in LoadURLsFromFile", zap.Error(err))
// 		return nil, err
// 	}
// 	defer file.Close()

// 	decoder := json.NewDecoder(file)
// 	var urlDataSlice []URLData
// 	for decoder.More() {
// 		var urlData URLData
// 		err := decoder.Decode(&urlData)
// 		if err != nil {
// 			logger.Log.Info("не смогли декодировать объект", zap.Error(err))
// 			return nil, err
// 		}
// 		urlDataSlice = append(urlDataSlice, urlData)
// 	}

// 	return urlDataSlice, nil
// }

func (us *URLShortener) CloseFileStorage() error {
	return us.fileStorage.Close() // Закрываем файл
}

// Метод для установки функции генерации идентификатора
func (us *URLShortener) SetGenerateIDFunc(fn func() string) {
	us.generateIDFunc = fn
}

func (us *URLShortener) GenerateUUID() int {
	us.uuidMu.Lock()
	defer us.uuidMu.Unlock()

	us.uuidCounter++
	return us.uuidCounter
}

func NewURLShortener(cfg *config.Config, storage storage.URLStorage, fileStorage *Producer) *URLShortener {
	us := &URLShortener{
		config:      cfg,
		Storage:     storage,
		fileStorage: fileStorage,
		uuidCounter: 0,
	}

	return us
}

func (us *URLShortener) APIShortenerURL(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	url := req.URL

	id := us.GenerateID()
	cfg := us.config
	shortenedURL := cfg.BaseURL + "/" + id

	err := us.Storage.SaveURL(id, url)
	if err != nil {
		http.Error(w, "error saving URL", http.StatusInternalServerError)
		return
	}

	urlData := &URLData{
		UUID:        us.GenerateUUID(),
		ShortURL:    shortenedURL,
		OriginalURL: string(url),
	}

	if err := us.fileStorage.SaveToFileURL(urlData); err != nil {
		logger.Log.Error("Error saving URL data in file", zap.Error(err))
		return
	}

	resp := models.Response{
		Result: shortenedURL,
	}
	jsonData, err := json.Marshal(&resp)
	if err != nil {
		http.Error(w, "error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
	}
}

func (us *URLShortener) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := us.config
	url, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	// Генерируем случайный идентификатор для сокращения URL
	id := us.GenerateID()
	shortenedURL := cfg.BaseURL + "/" + id

	// fmt.Printf("Received URL to save: id=%s, url=%s\n", id, string(url))
	// сохранение URL в мапу
	err = us.Storage.SaveURL(id, string(url))
	if err != nil {
		logger.Log.Error("Error saving URL", zap.Error(err))
		http.Error(w, "Error saving URL", http.StatusInternalServerError)
		return
	}

	urlData := &URLData{
		UUID:        us.GenerateUUID(),
		ShortURL:    shortenedURL,
		OriginalURL: string(url),
	}

	if err := us.fileStorage.SaveToFileURL(urlData); err != nil {
		logger.Log.Error("Error saving URL data in file", zap.Error(err))
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
		return
	}
}

func (us *URLShortener) RedirectURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	URLId := strings.TrimPrefix(r.URL.Path, "/")

	// fmt.Printf("redirect url %s \n", r.Host+r.URL.String())
	originalURL, err := us.Storage.GetURL(URLId)
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
	for i := 0; i < 10; i++ {
		index := rand.Intn(base)
		idBuilder.WriteByte(base62Alphabet[index])
	}

	return idBuilder.String()
}
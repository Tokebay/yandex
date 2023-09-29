package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Tokebay/yandex/config"
	"github.com/google/uuid"

	"github.com/Tokebay/yandex/internal/app/storage"
	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	"go.uber.org/zap"
)

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

	fmt.Printf("DSN %s; fileStorage %s \n", cfg.DSN, cfg.FileStoragePath)
	httpStatusCode := http.StatusCreated

	if cfg.DSN != "" {
		userID, err := us.GetNextUserID(w, r)
		fmt.Printf("shortener. user %d; err %s \n", userID, err)
		if err != nil {
			// w.WriteHeader(http.StatusBadRequest)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		pgStorage := us.Storage.(*storage.PostgreSQLStorage)
		fmt.Println("Save to DB")
		var mURL models.ShortenURL

		mURL.ShortURL = cfg.BaseURL + "/" + id
		mURL.OriginalURL = string(url)
		mURL.UserID = userID

		fmt.Printf("Received URL to save: id=%s, origURL %s, userID %d \n", mURL.ShortURL, mURL.OriginalURL, mURL.UserID)

		shortURL, err := pgStorage.InsertURL(mURL)
		if err != nil && shortURL == "" {
			httpStatusCode = http.StatusConflict
			shortURL, err = pgStorage.GetShortURL(string(url))
			if err != nil {
				logger.Log.Error("Error get Original URL", zap.Error(err))
				return
			}
		}
		shortenedURL = shortURL
	} else {
		urlData := &URLData{
			UUID:        us.GenerateUUID(),
			ShortURL:    shortenedURL,
			OriginalURL: string(url),
		}
		fmt.Println("Save to FILE")
		// сохранение URL в мапу
		err = us.Storage.SaveURL(id, string(url))
		if err != nil {
			logger.Log.Error("Error saving URL", zap.Error(err))
			http.Error(w, "Error saving URL", http.StatusInternalServerError)
			return
		}

		if err := us.fileStorage.SaveToFileURL(urlData); err != nil {
			logger.Log.Error("Error saving URL data in file", zap.Error(err))
			return
		}
	}

	fmt.Printf("Original URL: %s\n", url)
	fmt.Printf("Shortened URL: %s\n", shortenedURL)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(shortenedURL)))
	w.WriteHeader(httpStatusCode)
	_, err = w.Write([]byte(shortenedURL))

	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func (us *URLShortener) SaveToFile(urlData *URLData) error {

	if err := us.fileStorage.SaveToFileURL(urlData); err != nil {
		logger.Log.Error("Error saving URL data in file", zap.Error(err))
		return err
	}

	return nil
}

func (us *URLShortener) RedirectURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	URLId := strings.TrimPrefix(r.URL.Path, "/")
	cfg := us.config
	var originalURL string
	var err error
	if cfg.DSN != "" {
		shortURL := cfg.BaseURL + r.URL.Path

		// pgStorage, err := storage.NewPostgreSQLStorage(cfg.DSN)
		pgStorage := us.Storage.(*storage.PostgreSQLStorage)

		originalURL, err = pgStorage.GetURL(shortURL)
		if err != nil {
			logger.Log.Error("Error get row from DB", zap.Error(err))
		}

	} else {
		// fmt.Printf("redirect url %s \n", r.Host+r.URL.String())
		originalURL, err = us.Storage.GetURL(URLId)
		if err != nil {
			http.Error(w, "URL not found", http.StatusBadRequest)
			return
		}
	}
	// Выполняем перенаправление на оригинальный URL
	fmt.Printf("select originalURL %s", originalURL)
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

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

	userID, err := us.GetNextUserID(w, r)
	fmt.Printf("shortener. user %d; err %s \n", userID, err)
	if err != nil {
		// w.WriteHeader(http.StatusBadRequest)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := us.GenerateID()
	cfg := us.config

	shortenedURL := cfg.BaseURL + "/" + id

	httpStatusCode := http.StatusCreated
	if cfg.DSN != "" {
		pgStorage := us.Storage.(*storage.PostgreSQLStorage)

		var mURL models.ShortenURL
		mURL.OriginalURL = string(url)
		mURL.ShortURL = shortenedURL
		mURL.UserID = userID

		shortURL, err := pgStorage.InsertURL(mURL)
		if err != nil && shortURL == "" {
			httpStatusCode = http.StatusConflict
			shortURL, err = pgStorage.GetShortURL(string(url))
			if err != nil {
				logger.Log.Error("Error get Original URL", zap.Error(err))
				return
			}
		}
		shortenedURL = shortURL

	} else {
		urlData := &URLData{
			UUID:        us.GenerateUUID(),
			ShortURL:    shortenedURL,
			OriginalURL: string(url),
		}
		// сохранение URL в мапу
		err := us.Storage.SaveURL(id, string(url))
		if err != nil {
			logger.Log.Error("Error saving URL", zap.Error(err))
			http.Error(w, "Error saving URL", http.StatusInternalServerError)
			return
		}

		if err := us.fileStorage.SaveToFileURL(urlData); err != nil {
			logger.Log.Error("Error saving URL data in file", zap.Error(err))
			return
		}
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
	w.WriteHeader(httpStatusCode)

	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
	}
}

func (us *URLShortener) GetAllURLByUserID(w http.ResponseWriter, r *http.Request) {
	cfg := us.config
	if cfg.DSN != "" {
		userID, err := us.GetNextUserID(w, r)
		fmt.Printf("GetAllURLByUserID. user %d; err %s \n", userID, err)
		if err != nil {
			logger.Log.Error("GetAllURLByUserID. Error GetNextUserID", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Получите все URL пользователя из базы данных
		urls, err := us.GetUserURLs(userID)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			logger.Log.Error("Error getting user URLs from database", zap.Error(err))
			return
		}
		// Преобразуйте URLs в формат JSON и отправьте клиенту
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(urls)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}

}

func (us *URLShortener) GenerateID() string {
	// для тестов
	if us.generateIDFunc != nil {
		return us.generateIDFunc()
	}
	// Генерируем UUID
	u, err := uuid.NewRandom()
	if err != nil {
		// Обработка ошибки генерации UUID
		return ""
	}
	// Преобразуем UUID в строку, убираем дефисы и берем первые 10 символов
	id := strings.Replace(u.String(), "-", "", -1)
	if len(id) > 10 {
		id = id[:10]
	}

	return id
}

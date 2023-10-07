package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Tokebay/yandex/internal/app/storage"
	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	"go.uber.org/zap"
)

var ErrUserID = errors.New("error occured while getUserID")

func (us *URLShortener) BatchShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := us.config

	var req models.BatchShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var resp models.BatchShortenResponse
	httpStatusCode := http.StatusCreated

	var userID int
	if cfg.DSN != "" {
		userID, ErrUserID = us.GetNextUserID(w, r)
		if ErrUserID != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	var mURL models.ShortenURL
	mURL.UserID = userID

	for _, url := range req {
		id := us.GenerateID()
		shortenedURL := cfg.BaseURL + "/" + id

		// fmt.Printf("id %s ; origURL %s;  \n", url.CorrelationID, url.OriginalURL)
		if cfg.DSN != "" {

			mURL.ShortURL = shortenedURL
			mURL.OriginalURL = url.OriginalURL

			// pgStorage, err := storage.NewPostgreSQLStorage(cfg.DSN)
			pgStorage := us.Storage.(*storage.PostgreSQLStorage)
			shortURL, err := pgStorage.InsertURL(mURL)
			if err != nil && shortURL == "" {
				httpStatusCode = http.StatusConflict
				shortURL, err = pgStorage.GetShortURL(url.OriginalURL)
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
				OriginalURL: url.OriginalURL,
			}

			err := us.Storage.SaveURL(id, url.OriginalURL)
			if err != nil {
				http.Error(w, "Error saving URL", http.StatusInternalServerError)
				return
			}
			if err := us.fileStorage.SaveToFileURL(urlData); err != nil {
				logger.Log.Error("Error saving URL data in file", zap.Error(err))
				return
			}
		}
		resp = append(resp, struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}{
			CorrelationID: url.CorrelationID,
			ShortURL:      shortenedURL,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

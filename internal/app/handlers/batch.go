package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Tokebay/yandex/internal/app/storage"
	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	"go.uber.org/zap"
)

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

	for _, url := range req {
		id := us.GenerateID()
		shortenedURL := cfg.BaseURL + "/" + id

		// fmt.Printf("id %s ; origURL %s;  \n", url.CorrelationID, url.OriginalURL)
		if cfg.DSN != "" {
			pgStorage, err := storage.NewPostgreSQLStorage(cfg.DSN, us.dbPool)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			shortURL, err := pgStorage.InsertURL(shortenedURL, url.OriginalURL)
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

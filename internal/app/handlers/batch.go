package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Tokebay/yandex/internal/app/storage"
	"github.com/Tokebay/yandex/internal/logger"
	"go.uber.org/zap"
)

// request
type BatchShortenRequest []struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// response
type BatchShortenResponse []struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (us *URLShortener) BatchShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := us.config

	var req BatchShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var resp BatchShortenResponse
	for _, url := range req {
		id := us.GenerateID()
		shortenedURL := cfg.BaseURL + "/" + id

		// fmt.Printf("id %s ; origURL %s;  \n", url.CorrelationID, url.OriginalURL)
		if cfg.DSN != "" {
			// fmt.Println("api/batch Save in DB")
			pgStorage, err := storage.NewPostgreSQLStorage(cfg.DSN)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			err = pgStorage.SaveURL(shortenedURL, url.OriginalURL)
			if err != nil {
				http.Error(w, "Error saving URL in DB", http.StatusInternalServerError)
				return
			}

		} else {
			urlData := &URLData{
				UUID:        us.GenerateUUID(),
				ShortURL:    shortenedURL,
				OriginalURL: url.OriginalURL,
			}

			// fmt.Println("api/batch Save in FILE")
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
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

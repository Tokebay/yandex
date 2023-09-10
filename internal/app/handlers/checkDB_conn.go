package handlers

import (
	"database/sql"
	"net/http"

	"github.com/Tokebay/yandex/internal/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// проверяем соединение с БД
func (us *URLShortener) CheckDBConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := us.GetDB()
	if err != nil {
		logger.Log.Error("Error getDB", zap.Error(err))
	}
	w.WriteHeader(http.StatusOK)
	// _, err = w.Write([]byte("success"))
}

func (us *URLShortener) GetDB() (*sql.DB, error) {

	dbConnString := us.config.DataBaseConnString

	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		logger.Log.Error("Error open connection with DB", zap.Error(err))

	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Log.Error("Error establishing connection with DB", zap.Error(err))
		return nil, err
	}

	return db, nil
}

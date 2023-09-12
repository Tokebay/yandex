package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (us *URLShortener) CreateTable() (*gorm.DB, error) {

	db, err := us.PostgresInit()
	if err != nil {
		logger.Log.Error("Error connect to DB", zap.Error(err))
		return nil, err
	}
	db.AutoMigrate(&models.ShortenURL{})
	return db, nil
}

func (us *URLShortener) GetOriginDBURL(shortenURL string) (string, error) {
	db, err := us.PostgresInit()
	if err != nil {
		logger.Log.Error("Error connect to DB", zap.Error(err))
		return "", err
	}
	// Get  matched record
	var url models.ShortenURL
	if err := db.Select("original_url").Where("short_url = ?", &shortenURL).First(&url).Error; err != nil {
		return "", err
	}

	return url.OriginalURL, nil
}

// проверяем соединение с БД
func (us *URLShortener) CheckDBConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := us.GetDB()
	if err != nil {
		logger.Log.Error("Error connect to DB", zap.Error(err))
	}

	w.WriteHeader(http.StatusOK)
	// _, err = w.Write([]byte("success"))
}

func (us *URLShortener) GetDB() (*sql.DB, error) {

	dbConnString := us.config.DataBaseConnString
	fmt.Printf("dsn %s \n", dbConnString)
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

func (us *URLShortener) PostgresInit() (*gorm.DB, error) {
	dsn := us.config.DataBaseConnString

	// fmt.Printf("dsn %s \n", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Error("Error establishing connection with DB", zap.Error(err))
		return nil, err
	}

	return db, err
}

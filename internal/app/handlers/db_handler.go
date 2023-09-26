package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// проверяем соединение с БД
func (us *URLShortener) CheckDBConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := us.GetDB()
	if err != nil {
		logger.Log.Error("Error connect to DB", zap.Error(err))
		http.Error(w, "Error connect to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = CreateShortenedURLTable(db)
	if err != nil {
		logger.Log.Error("Error create table", zap.Error(err))
		http.Error(w, "Error create table", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (us *URLShortener) GetDB() (*sql.DB, error) {

	dbConnString := us.config.DSN
	fmt.Printf("dsn %s \n", dbConnString)
	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		logger.Log.Error("Error open connection with DB", zap.Error(err))

	}
	return db, nil
}

func CreateShortenedURLTable(db *sql.DB) error {
	// Создаем таблицу shortened_urls, если она еще не существует
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS public.shorten_urls
	(
		uuid SERIAL,
		short_url text COLLATE pg_catalog."default",
		original_url text COLLATE pg_catalog."default"
	)`)
	if err != nil {
		return err
	}
	return nil
}

func (us *URLShortener) InsertData(db *sql.DB, url *models.ShortenURL) error {

	_, err := db.Exec(`
        INSERT INTO shorten_urls (short_url, original_url)
        VALUES ($1, $2)`,
		url.ShortURL, url.OriginalURL)

	return err
}

func (us *URLShortener) SelectURLData(db *sql.DB, shortURL string) (string, error) {

	fmt.Printf("shortURL %s \n", shortURL)
	var url models.ShortenURL
	row := db.QueryRow("SELECT original_url FROM shorten_urls where short_url=$1", shortURL)
	err := row.Scan(&url.OriginalURL)
	if err != nil {
		logger.Log.Error("Error select row from DB", zap.Error(err))
		return "", err
	}
	// fmt.Printf("selectURL url.OriginalURL %s \n", url.OriginalURL)
	return url.OriginalURL, nil
}

func (us *URLShortener) SaveToDB(shortenURL *models.ShortenURL) (*sql.DB, error) {
	//ping DB
	db, err := us.GetDB()
	if err != nil {
		logger.Log.Error("Error connect to DB", zap.Error(err))
	}
	err = us.InsertData(db, shortenURL)
	if err != nil {
		logger.Log.Info("Error insert data to table", zap.Error(err))
		return nil, err
	}

	return db, nil
}

// type Postgres struct {
// 	store *sql.DB
// }

// func (us *URLShortener) MigrateTable(*Postgres, error) (*Postgres, error) {

// 	dsn := us.config.DSN
// 	db, err := goose.OpenDBWithDriver("pgx", dsn)
// 	if err != nil {
// 		logger.Log.Error("Error open conn", zap.Error(err))
// 	}

// 	err = goose.Up(db, "./migrations")
// 	if err != nil {
// 		logger.Log.Error("Error goose sUP", zap.Error(err))
// 	}

// 	return &Postgres{
// 		store: db,
// 	}, nil

// }

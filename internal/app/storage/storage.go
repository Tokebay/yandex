package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type URLStorage interface {
	SaveURL(id, url string) error
	GetURL(id string) (string, error)
}

type MapStorage struct {
	mapping map[string]string
	mu      sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		mapping: make(map[string]string),
	}
}

func (ms *MapStorage) SaveURL(shortenURL, originalURL string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	//fmt.Printf("save id %s; url %s \n", id, url)
	ms.mapping[shortenURL] = originalURL
	// fmt.Printf("Saved URL: id=%s, url=%s\n", id, url)
	return nil
}

func (ms *MapStorage) GetURL(shortenURL string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	url, ok := ms.mapping[shortenURL]
	fmt.Printf("getURL %s; url %s \n", shortenURL, url)
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}

type PostgreSQLStorage struct {
	db *sql.DB
}

func (s *PostgreSQLStorage) Close() error {
	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			logger.Log.Error("Error closing database connection", zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *PostgreSQLStorage) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func NewPostgreSQLStorage(dsn string) (*PostgreSQLStorage, error) {
	// Выполнить миграции
	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		logger.Log.Error("Error open conn", zap.Error(err))
		return nil, err
	}
	err = goose.Up(db, "./database/migration")
	if err != nil {
		logger.Log.Error("Error goose UP", zap.Error(err))
		return nil, err
	}

	// Вернуть созданный объект PostgreSQLStorage
	return &PostgreSQLStorage{db: db}, nil
}

var ErrAlreadyExistURL = errors.New("URLAlreadyExist")

func (s *PostgreSQLStorage) SaveURL(shortURL string, origURL string) error {
	// ctx := context.Background()

	// Запрос использует RETURNING, поэтому нам нужно предоставить переменную для получения результата
	var returnedShortURL string

	// Выполним запрос с помощью pgx
	err := s.db.QueryRow(`
		 INSERT INTO shorten_urls (short_url, original_url)
		 VALUES ($1, $2)
		 ON CONFLICT (original_url) DO NOTHING
		 RETURNING short_url`, shortURL, origURL).Scan(&returnedShortURL)

	if err != nil {
		if err == pgx.ErrNoRows { // если ON CONFLICT не сработал и ни одна строка не вернулась
			fmt.Println("rowsAffected 0")
			return ErrAlreadyExistURL
		}
		logger.Log.Error("Error insert URL to table", zap.Error(err))
		return err
	}

	return nil
}

func (s *PostgreSQLStorage) InsertUser() (int, error) {

	var userID int

	err := s.db.QueryRow(`INSERT INTO users_links DEFAULT VALUES RETURNING user_id`).Scan(&userID)

	if err != nil {
		logger.Log.Error("Error Insert Users", zap.Error(err))
		return 0, err
	}

	return userID, nil
}

func (s *PostgreSQLStorage) InsertURL(url models.ShortenURL) (string, error) {
	// ctx := context.Background()

	var existingShortURL string

	err := s.db.QueryRow(`INSERT INTO shorten_urls (short_url, original_url,user_id)
	    VALUES ($1, $2, $3)
	    ON CONFLICT (original_url) DO NOTHING
	    RETURNING short_url`, url.ShortURL, url.OriginalURL, url.UserID).Scan(&existingShortURL)

	if err != nil {
		logger.Log.Error("Error Insert URL to table", zap.Error(err))
		return "", err
	}

	return existingShortURL, nil
}

// GetURL получает URL из PostgreSQL
func (s *PostgreSQLStorage) GetURL(shortURL string) (string, error) {
	// ctx := context.Background()
	var url models.ShortenURL
	row := s.db.QueryRow("SELECT original_url FROM shorten_urls where short_url=$1 and is_deleted != true", shortURL)
	err := row.Scan(&url.OriginalURL)
	if err != nil {
		logger.Log.Error("No row selected from table", zap.Error(err))
		return "", err
	}

	return url.OriginalURL, nil
}

func (s *PostgreSQLStorage) GetShortURL(origURL string) (string, error) {
	// ctx := context.Background()
	var url models.ShortenURL
	err := s.db.QueryRow("SELECT short_url FROM shorten_urls WHERE original_url = $1", origURL).Scan(&url.ShortURL)
	if err != nil {
		logger.Log.Error("Error in GetOrigURL. short_url", zap.Error(err))
		return "", err
	}
	return url.ShortURL, nil
}

func (s *PostgreSQLStorage) indexExists(indexName string) (bool, error) {
	// ctx := context.Background()
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE indexname = $1
		)`
	var exists bool
	err := s.db.QueryRow(query, indexName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *PostgreSQLStorage) MarkURLAsDeleted(userID int, url string) error {
	// Обновление записи в базе данных для удаления URL, учитывая userID
	fmt.Printf("MarkURLAsDeleted userID %d, url %s \n", userID, url)
	query := "UPDATE shorten_urls SET is_deleted = true WHERE user_id = $1 AND short_url = $2"
	_, err := s.db.Exec(query, userID, url)
	if err != nil {
		logger.Log.Error("error update shorten_urls", zap.Error(err))
		return err
	}
	return nil
}

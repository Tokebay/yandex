package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
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

// NewPostgreSQLStorage новое PostgreSQL хранилище с заданным DSN
func NewPostgreSQLStorage(dsn string) (*PostgreSQLStorage, error) {

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Log.Error("Error open connection to DB", zap.Error(err))
		return nil, err
	}

	return &PostgreSQLStorage{db: db}, nil
}

// SaveURL сохраняет URL в PostgreSQL
func (s *PostgreSQLStorage) SaveURL(shortURL string, origURL string) error {
	// сохранение URL в PostgreSQL

	_, err := s.db.Exec(`
        INSERT INTO shorten_urls (short_url, original_url)
        VALUES ($1, $2)`,
		shortURL, origURL)
	if err != nil {
		logger.Log.Error("Error insert URL", zap.Error(err))
		return err
	}
	return nil
}

// GetURL получает URL из PostgreSQL
func (s *PostgreSQLStorage) GetURL(shortURL string) (string, error) {
	// получение URL из PostgreSQL
	var url models.ShortenURL
	row := s.db.QueryRow("SELECT original_url FROM shorten_urls where short_url=$1", shortURL)
	err := row.Scan(&url.OriginalURL)
	if err != nil {
		logger.Log.Error("No row selected from table", zap.Error(err))
		return "", err
	}

	return url.OriginalURL, nil
}

func (s *PostgreSQLStorage) CreateTable() error {
	// Создание таблицы в PostgreSQL
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS public.shorten_urls
	(
		uuid SERIAL,
		short_url text NOT NULL,
		original_url text NOT NULL
	)`)
	if err != nil {
		logger.Log.Error("Error occured create table", zap.Error(err))
		return err
	}
	return nil
}

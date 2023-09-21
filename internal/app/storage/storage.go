package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Tokebay/yandex/internal/logger"
	"github.com/Tokebay/yandex/internal/models"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
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
	db *pgxpool.Pool
}

func NewPostgreSQLStorage(dsn string, dbPool *pgxpool.Pool) (*PostgreSQLStorage, error) {
	return &PostgreSQLStorage{db: dbPool}, nil
}

var ErrAlreadyExistURL = errors.New("URLAlreadyExist")

func (s *PostgreSQLStorage) SaveURL(shortURL string, origURL string) error {
	// сохранение URL в PostgreSQL

	// Сперва создадим контекст
	ctx := context.Background()

	// Запрос использует RETURNING, поэтому нам нужно предоставить переменную для получения результата
	var returnedShortURL string

	// Выполним запрос с помощью pgx
	err := s.db.QueryRow(ctx, `
		 INSERT INTO shorten_urls (short_url, original_url)
		 VALUES ($1, $2)
		 ON CONFLICT (original_url) DO NOTHING
		 RETURNING short_url
	 `, shortURL, origURL).Scan(&returnedShortURL)

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

func (s *PostgreSQLStorage) InsertURL(shortURL string, origURL string) (string, error) {
	ctx := context.Background()

	var existingShortURL string

	err := s.db.QueryRow(ctx, `
	    INSERT INTO shorten_urls (short_url, original_url)
	    VALUES ($1, $2)
	    ON CONFLICT (original_url) DO NOTHING
	    RETURNING short_url`, shortURL, origURL).Scan(&existingShortURL)

	if err != nil {
		logger.Log.Error("Error Insert URL to table", zap.Error(err))
		return "", err
	}

	return existingShortURL, nil
}

// GetURL получает URL из PostgreSQL
func (s *PostgreSQLStorage) GetURL(shortURL string) (string, error) {
	ctx := context.Background()
	// получение URL из PostgreSQL
	var url models.ShortenURL
	row := s.db.QueryRow(ctx, "SELECT original_url FROM shorten_urls where short_url=$1", shortURL)
	err := row.Scan(&url.OriginalURL)
	if err != nil {
		logger.Log.Error("No row selected from table", zap.Error(err))
		return "", err
	}

	return url.OriginalURL, nil
}

func (s *PostgreSQLStorage) GetShortURL(origURL string) (string, error) {
	ctx := context.Background()
	var url models.ShortenURL
	err := s.db.QueryRow(ctx, "SELECT short_url FROM shorten_urls WHERE original_url = $1", origURL).Scan(&url.ShortURL)
	if err != nil {
		logger.Log.Error("Error in GetOrigURL. short_url", zap.Error(err))
		return "", err
	}
	return url.ShortURL, nil
}

func (s *PostgreSQLStorage) CreateTable() error {
	// Создание таблицы в PostgreSQL
	ctx := context.Background()
	_, err := s.db.Exec(ctx,
		`CREATE TABLE IF NOT EXISTS public.shorten_urls
		(
			uuid SERIAL,
			short_url text NOT NULL,
			original_url text NOT NULL
		)`)

	if err != nil {
		logger.Log.Error("Error occured create table", zap.Error(err))
		return err
	}

	indexName := "original_url_index"

	// Проверяем существование индекса
	exists, err := s.indexExists(indexName)
	if err != nil {
		logger.Log.Error("Error create index", zap.Error(err))
		return err
	}
	fmt.Println("indexExist", exists)
	// Если индекс не существует, создаем его
	if !exists {
		createIndexSQL := fmt.Sprintf("CREATE UNIQUE INDEX %s ON shorten_urls (original_url)", indexName)
		_, err := s.db.Exec(ctx, createIndexSQL)
		if err != nil {
			logger.Log.Error("Error create index", zap.Error(err))
			return err
		}

	}

	return nil
}

func (s *PostgreSQLStorage) indexExists(indexName string) (bool, error) {
	ctx := context.Background()
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE indexname = $1
		)`

	var exists bool
	err := s.db.QueryRow(ctx, query, indexName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

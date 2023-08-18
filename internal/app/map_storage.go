package app

import (
	"errors"
	"sync"
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

func (ms *MapStorage) SaveURL(id, url string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.mapping[id] = url

	return nil
}

func (ms *MapStorage) GetURL(id string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	url, ok := ms.mapping[id]

	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}

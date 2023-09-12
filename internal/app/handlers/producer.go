package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tokebay/yandex/internal/logger"
	"go.uber.org/zap"
)

type Producer struct {
	file     *os.File
	encoder  *json.Encoder
	filePath string
	buffer   []URLData
}

func NewProducer(filePath string) (*Producer, error) {

	dir := filepath.Dir(filePath)
	fmt.Printf("dirName %s\n", dir)

	err := os.MkdirAll("/tmp", 0755)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("err %s", err)
		return nil, err
	}

	currentDir, err := os.Getwd()
	fmt.Printf("currDir: %s\n", currentDir)
	if err != nil {
		logger.Log.Info("Error getting current working directory", zap.Error(err))
	}

	return &Producer{
		file:     file,
		encoder:  json.NewEncoder(file),
		filePath: filePath,
	}, nil
}

func (p *Producer) SaveToFileURL(urlData *URLData) error {
	// Загрузка существующих данных
	existingData, err := p.LoadInitialData()
	if err != nil {
		return err
	}

	// Append the new URLData to the existing data
	existingData = append(existingData, *urlData)

	for _, u := range existingData {
		parts := strings.Split(u.ShortURL, "/")
		URLId := parts[len(parts)-1]
		fmt.Printf("shortURL %s; partID %s; origURL %s \n", u.ShortURL, URLId, u.OriginalURL)
	}

	// Write the updated data back to the file
	if err := p.WriteToFile(existingData); err != nil {
		logger.Log.Error("Error saving URL in file", zap.Error(err))
		return err
	}

	return nil
}

func (p *Producer) LoadInitialData() ([]URLData, error) {
	fmt.Printf("p.filePath loadFromFile %s: \n", p.filePath)
	file, err := os.OpenFile(p.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Error("Error opening file for reading", zap.Error(err))
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var urlDataSlice []URLData
	for decoder.More() {
		var urlData URLData
		err := decoder.Decode(&urlData)
		if err != nil {
			logger.Log.Error("Error decoding data from file", zap.Error(err))
			return nil, err
		}
		urlDataSlice = append(urlDataSlice, urlData)
	}

	return urlDataSlice, nil
}

func (p *Producer) WriteToFile(urlData []URLData) error {
	file, err := os.OpenFile(p.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Log.Error("Error opening file for writing", zap.Error(err))
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, data := range urlData {
		if err := encoder.Encode(data); err != nil {
			logger.Log.Error("Error encoding data to file", zap.Error(err))
			return err
		}
	}

	return nil
}

func (p *Producer) Close() error {
	if err := p.file.Close(); err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}
	return nil

}

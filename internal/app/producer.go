package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		buffer:   nil,
	}, nil
}

func (p *Producer) SaveToFileURL(urlData *URLData) error {
	p.buffer = append(p.buffer, *urlData)
	if err := p.Flush(); err != nil {
		logger.Log.Error("Error saving URL in file", zap.Error(err))
		return err
	}
	return nil
}

func (p *Producer) Flush() error {
	fmt.Printf("p.filePath %s\n", p.filePath)

	for _, urlData := range p.buffer {
		if err := p.encoder.Encode(urlData); err != nil {
			logger.Log.Error("Error encoding data", zap.Error(err))
			return err
		}
	}

	p.buffer = nil

	return nil
}

func (p *Producer) Close() error {
	if err := p.Flush(); err != nil {
		return fmt.Errorf("error flushing buffer: %w", err)
	}
	if err := p.file.Close(); err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}
	return nil

}

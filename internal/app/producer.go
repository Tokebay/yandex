package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filePath string) (*Producer, error) {

	dir := filepath.Dir(filePath)
	// fmt.Printf("dir: %s\n", dir)
	err := os.MkdirAll(dir, 0755) //create the directory and give it required permissions
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("filePath %s \n", filePath)

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteEvent(urlData *URLData) error {
	return p.encoder.Encode(&urlData)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

package app

import (
	"encoding/json"
	"fmt"
	"os"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	fmt.Printf("fileName %s \n", fileName)

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

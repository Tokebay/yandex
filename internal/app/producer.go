package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Producer struct {
	file     *os.File
	encoder  *json.Encoder
	filePath string
	buffer   []URLData
}

func NewProducer(filePath string) (*Producer, error) {

	fmt.Printf("dirPath %s \n", filePath)
	splitDir := strings.Split(filePath, "/")
	dirName := splitDir[0]
	fmt.Printf("dirName %s\n", dirName)

	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return nil, err
	}

	// fmt.Println("filePath ", filePath)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:     file,
		encoder:  json.NewEncoder(file),
		filePath: filePath,
		buffer:   nil,
	}, nil
}

func (p *Producer) WriteEvent(urlData *URLData) error {
	p.buffer = append(p.buffer, *urlData)
	return p.Flush()
}

func (p *Producer) Flush() error {
	// return p.encoder.Encode(&p.buffer)
	file, err := os.OpenFile(p.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode and append the buffer contents to the file
	encoder := json.NewEncoder(file)
	for _, urlData := range p.buffer {
		if err := encoder.Encode(urlData); err != nil {
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

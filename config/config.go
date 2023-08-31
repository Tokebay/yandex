package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	ServerPort      int
	FileStoragePath string
}

func NewConfig() *Config {
	config := &Config{}
	config.parseEnv()

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.IntVar(&config.ServerPort, "p", 8080, "HTTP server port")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/short-url-db.json", "Path to FILE_STORAGE_PATH")

	flag.Parse()

	return config
}

func (c *Config) parseEnv() {
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		c.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		c.BaseURL = envBaseURL
	}

	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		port, err := strconv.Atoi(serverPort)
		fmt.Println("port ", port, serverPort)
		if err == nil {
			c.ServerPort = port
		}
	}

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		fmt.Println("envFilePath ", envFilePath)
		c.FileStoragePath = envFilePath
	}
}

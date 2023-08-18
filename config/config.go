package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	ServerPort    int
}

func NewConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.ServerAddress, "a", "localhost", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")

	flag.Parse()
	config.parseEnv()
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
		if err == nil {
			c.ServerPort = port
		}
	} else {
		c.ServerPort = 8080 // Установка значения по умолчанию
	}
}

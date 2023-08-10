package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	ServerPort    int
}

func NewConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
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
		c.ServerPort = 8080 // Значение по умолчанию, если серверный порт не установлен
	}
}

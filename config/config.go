package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func ParseFlags() *Config {
	config := &Config{}
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.Parse()
	return config
}

func (c *Config) ParseEnv() {
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		c.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		c.BaseURL = envBaseURL
	}
}

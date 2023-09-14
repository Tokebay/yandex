package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	ServerPort      string
	FileStoragePath string
	DSN             string
	DataBaseConn    DataBase
	UseDatabase     bool // использования базы данных
}

type DataBase struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.StringVar(&config.ServerPort, "p", "8080", "HTTP server port")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/short-url-db.json", "Path to FILE_STORAGE_PATH")

	// db := &DataBase{
	// 	DBName:   "postgres",
	// 	Host:     "127.0.0.1",
	// 	Port:     5432,
	// 	User:     "postgres",
	// 	Password: "postgres",
	// }

	// postgresConnString := fmt.Sprintf("host=%s port=%d user=%s "+
	// 	"password=%s dbname=%s sslmode=disable",
	// 	db.Host, db.Port, db.User, db.Password, db.DBName)

	// flag.StringVar(&config.DataBaseConnString, "d", postgresConnString, "Database connection string DSN")

	flag.StringVar(&config.DSN, "d", "", "Database DSN") // Добавляем флаг для строки подключения к БД
	flag.BoolVar(&config.UseDatabase, "use-db", false, "Use PostgreSQL database")

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
		c.ServerPort = serverPort

	}

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		c.FileStoragePath = envFilePath
	}

	if envDBDSN := os.Getenv("DATABASE_DSN"); envDBDSN != "" {
		c.DSN = envDBDSN
		c.UseDatabase = true
	}
}

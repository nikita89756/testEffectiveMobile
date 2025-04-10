package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database
	Server
	Cache
	LogLevel string
}

type Database struct {
	DatabaseConnection string
	MigrationDir       string
	DBTimeout          time.Duration
}

type Cache struct {
	Address  string
	Password string
	Db       int
}

type Server struct {
	Host string
	Port string
}

// type APIs{

// }

func InitConfig() Config {
	err := godotenv.Load()
	if err != nil {
		db := os.Getenv("CACHE_DB")
		dbint, err := strconv.Atoi(db)
		if err != nil {
			log.Fatal("Не удалось преобразовать значение CACHE_DB в int:", err)
		}
		cfg := Config{
			Database: Database{
				DatabaseConnection: os.Getenv("DATABASE_CONNECTION"),
				MigrationDir:       os.Getenv("MIGRATION_DIR"),
				DBTimeout:          5 * time.Second,
			},
			Server: Server{
				Host: os.Getenv("SERVER_HOST"),
				Port: os.Getenv("SERVER_PORT"),
			},
			Cache: Cache{
				Address:  os.Getenv("CACHE_ADDRESS"),
				Password: os.Getenv("CACHE_PASSWORD"),
				Db:       dbint,
			},
			LogLevel: os.Getenv("LOG_LEVEL"),
		}
		if cfg.Database.DatabaseConnection == "" {
			log.Fatal("Не указана строка подключения к базе данных")
		}
		return cfg
	}
	db := os.Getenv("CACHE_DB")
	dbint, err := strconv.Atoi(db)
	if err != nil {
		log.Fatal("Не удалось преобразовать значение HASH_DB в int:", err)
	}
	cfg := Config{
		Database: Database{
			DatabaseConnection: getEnv("DATABASE_CONNECTION", ""),
			MigrationDir:       getEnv("MIGRATION_DIR", ""),
			DBTimeout:          5 * time.Second,
		},
		Cache: Cache{
			Address:  os.Getenv("CACHE_ADDRESS"),
			Password: os.Getenv("CACHE_PASSWORD"),
			Db:       dbint,
		},
		Server: Server{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	if cfg.Database.DatabaseConnection == "" {
		log.Fatal("Не указана строка подключения к базе данных")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

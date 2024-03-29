package config

import (
	"fmt"
	"time"
)

type Config struct {
	Secret          string
	MaxSessionCount int
	HTTPServer      HTTPServer
	Mongodb         Mongodb
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type HTTPServer struct {
	Address string
}

type Mongodb struct {
	Host     string
	Port     int
	Database string
	URI      string
}

func MustLoad() *Config {
	// из-за ограниченного стека используемых технологий не использую godotenv и cleanenv для заполнения конфига
	cfg := &Config{
		Secret:          "get_secret_from_env",
		MaxSessionCount: 3,
		HTTPServer:      HTTPServer{Address: "localhost:8080"},
		Mongodb: Mongodb{
			Host:     "localhost",
			Port:     27017,
			Database: "medods_test_task",
		},
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 60 * 24 * 60 * time.Minute,
	}

	cfg.Mongodb.URI = fmt.Sprintf("mongodb://%s:%d", cfg.Mongodb.Host, cfg.Mongodb.Port)

	return cfg
}

// YmFiMDgzMDgtZTQ3Yy00YzBkLWIyNTgtZTJlYWFmMDA3NDJmLTgxMzMzOGQ4YzllZGNi

package config

import "fmt"

type Config struct {
	Secret     string
	HTTPServer HTTPServer
	Mongodb    Mongodb
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
		Secret:     "get_secret_from_env",
		HTTPServer: HTTPServer{Address: "localhost:8080"},
		Mongodb: Mongodb{
			Host:     "localhost",
			Port:     27017,
			Database: "medods_test_task",
		},
	}

	cfg.Mongodb.URI = fmt.Sprintf("mongodb://%s:%s", cfg.Mongodb.Host, cfg.Mongodb.Port)

	return cfg
}

package confiq

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Address string
	BaseURL string
}

func LoadConfig() *Config {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", "", "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "", "Базовый адрес результирующего сокращенного URL")

	flag.Parse()
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.Address = envAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	if cfg.Address == "" {
		cfg.Address = "localhost:8080"
	}

	if cfg.BaseURL == "" {
		if strings.HasPrefix(cfg.Address, ":") {
			cfg.BaseURL = fmt.Sprintf("http://localhost%s", cfg.Address)
		} else {
			cfg.BaseURL = fmt.Sprintf("http://%s", cfg.Address)
		}
	}

	return &cfg
}

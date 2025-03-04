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

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "", "Базовый адрес сокращенного URL")
	flag.Parse()

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		if !strings.Contains(cfg.Address, ":") {
			cfg.Address = "localhost:" + envPort
		} else {
			cfg.Address = strings.Split(cfg.Address, ":")[0] + ":" + envPort
		}
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://%s", cfg.Address)
	}

	return &cfg
}

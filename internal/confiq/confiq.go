package confiq

import (
	"flag"
	"fmt"
	"net"
	"os"
)

type Config struct {
	Address string
	BaseURL string
}

func LoadConfig() *Config {
	var cfg Config

	defaultAddress := "localhost:8080"

	flag.StringVar(&cfg.Address, "a", defaultAddress, "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "", "Базовый адрес сокращенного URL")
	flag.Parse()

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		host, _, err := net.SplitHostPort(cfg.Address)
		if err != nil {
			host = "localhost"
		}
		cfg.Address = net.JoinHostPort(host, envPort)
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://%s", cfg.Address)
	}

	return &cfg
}

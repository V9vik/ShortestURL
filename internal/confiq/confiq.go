package confiq

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	Address string
	BaseURL string
}

func LoadConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for short links")
	flag.Parse()

	if cfg.BaseURL == "" {
		if strings.HasPrefix(cfg.Address, ":") {
			cfg.BaseURL = fmt.Sprintf("http://localhost%s", cfg.Address)
		} else {
			cfg.BaseURL = fmt.Sprintf("http://%s", cfg.Address)
		}
	}

	return &cfg
}

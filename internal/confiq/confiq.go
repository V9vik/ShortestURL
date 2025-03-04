package confiq

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	Address string
	BaseURL string
}

func LoadConfig() *Config {
	var cfg Config
	cfg.Address = os.Getenv("SERVER_ADDRESS")
	cfg.BaseURL = os.Getenv("BASE_URL")
	if cfg.Address == "" {

		flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP server address")
		flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for short links")
		flag.Parse()

		if cfg.BaseURL == "" {
			if strings.HasPrefix(cfg.Address, ":") {
				cfg.BaseURL = "http://localhost%s" + cfg.Address
			} else {
				cfg.BaseURL = "http://" + cfg.Address
			}
		}

		return &cfg
	} else {
		return &cfg
	}
}

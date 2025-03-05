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

func NewConfigFromFlags(flagSet *flag.FlagSet) *Config {
	var cfg Config

	defaultAddress := "localhost:8080"

	flagSet.StringVar(&cfg.Address, "a", defaultAddress, "Адрес запуска HTTP-сервера")
	flagSet.StringVar(&cfg.BaseURL, "b", "", "Базовый адрес сокращенного URL")
	flagSet.Parse(os.Args[1:])

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

func LoadConfig() *Config {
	return NewConfigFromFlags(flag.CommandLine)
}

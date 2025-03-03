package confiq

import (
	"flag"
	"fmt"
)

type Config struct {
	URLBase string
	Port    string
}

func LoadConfig() *Config {
	var address string
	var urlBase string

	flag.StringVar(&address, "a", "localhost:8080", "server address (host:port)")
	flag.StringVar(&urlBase, "b", "", "base URL for short links")
	flag.Parse()

	if urlBase == "" {
		urlBase = fmt.Sprintf("http://%s/", address)
	}

	return &Config{
		URLBase: urlBase,
		Port:    address,
	}
}

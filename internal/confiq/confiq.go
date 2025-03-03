package confiq

import (
	"flag"
)

type Config struct {
	UrlBase string
	Port    string
}

func LoadConfig() *Config {
	var port string
	var urlBase string
	flag.StringVar(&port, "a", "localhost:8080", "port to run server on")
	flag.StringVar(&urlBase, "b", "", "url to shortest URL")
	flag.Parse()
	if urlBase == "" {
		urlBase += "http://" + port
	}

	return &Config{
		UrlBase: urlBase,
		Port:    port,
	}

}

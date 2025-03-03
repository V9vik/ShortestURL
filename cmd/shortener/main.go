package main

import (
	"crypto/rand"
	"encoding/base32"
	"github.com/V9vik/ShortestURL.git/internal/confiq"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type urlStore struct {
	mu    sync.RWMutex
	store map[string]string
}

var (
	store = urlStore{
		store: make(map[string]string),
	}
)

func generateID() (string, error) {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

func handlerPost(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid content type"})
		return
	}

	url, err := io.ReadAll(c.Request.Body)
	if err != nil || len(url) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Empty body"})
		return
	}

	var shortID string
	for {
		var err error
		shortID, err = generateID()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}

		store.mu.Lock()
		if _, exists := store.store[shortID]; !exists {
			store.store[shortID] = string(url)
			store.mu.Unlock()
			break
		}
		store.mu.Unlock()
	}

	c.String(http.StatusCreated, "http://localhost:8080/%s", shortID)
}
func handlerGet(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	store.mu.RLock()
	longURL, exists := store.store[id]
	store.mu.RUnlock()

	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, longURL)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	cfg := confiq.LoadConfig()
	router := gin.Default()
	router.POST("/", handlerPost)
	router.GET("/:id", handlerGet)

	log.Println("Server start in:", cfg.UrlBase)
	log.Println(cfg.UrlBase)
	log.Println(cfg.Port)
	if err := router.Run(cfg.Port); err != nil {
		log.Fatal(err)
	}
}

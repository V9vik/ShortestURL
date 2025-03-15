package main

import (
	"crypto/rand"
	"encoding/base32"
	"github.com/V9vik/ShortestURL.git/internal/confiq"
	"github.com/V9vik/ShortestURL.git/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
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
	IDLength = 8
)

func generateID() (string, error) {
	buf := make([]byte, IDLength)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

func handlerPost(c *gin.Context, address string) {
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

	c.String(http.StatusCreated, address+"/%s", shortID)
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

	logger, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	cfg := config.LoadConfig()

	router := gin.Default()

	router.Use(ginLogger(sugar))

	router.POST("/", func(c *gin.Context) {
		handlerPost(c, cfg.BaseURL)
	})

	router.GET("/:id", handlerGet)

	sugar.Info("Server starting on:", cfg.BaseURL)
	if err := router.Run(cfg.Address); err != nil {
		sugar.Fatal(err)
	}
}

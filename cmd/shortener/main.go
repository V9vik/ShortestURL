package main

import (
	"crypto/rand"
	"encoding/base32"
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

var store = urlStore{
	store: make(map[string]string),
}

func generateID() (string, error) {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

func handlerPost(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return
	}

	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		http.Error(res, "Invalid content type", http.StatusBadRequest)
		return
	}

	url, err := io.ReadAll(req.Body)
	if err != nil || len(url) == 0 {
		http.Error(res, "Empty body", http.StatusBadRequest)
		return
	}

	var shortID string
	for {
		var err error
		shortID, err = generateID()
		if err != nil {
			http.Error(res, "Internal error", http.StatusInternalServerError)
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

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, "http://localhost:8080/"+shortID)
}

func handlerGet(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/")
	if id == "" {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	longURL, exists := store.store[id]
	store.mu.RUnlock()

	if !exists {
		http.NotFound(res, req)
		return
	}

	res.Header().Set("Location", longURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handlerPost(w, r)
		} else {
			handlerGet(w, r)
		}
	})

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed: ", err)
	}
}

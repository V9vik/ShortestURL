package main

import (
	"encoding/base32"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
)

type urlStore struct {
	mu    sync.Mutex
	store map[string]string
}

var (
	store     urlStore
	idCounter = 9
)

func generateID() (string, error) {
	buf := make([]byte, idCounter)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(buf), nil
}

func handlerPost(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return

	}

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(req.Header.Get("Content-Type"), "text/plain") {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := io.ReadAll(req.Body)
	if err != nil || len(url) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var shortID string
	for {
		shortID, err = generateID()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		store.mu.Lock()
		_, exists := store.store[shortID]
		store.mu.Unlock()
		if !exists {
			break
		}
	}
	store.store[shortID] = string(url)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, "http://localhost:8080/"+shortID+"\n")
}

func handlerGet(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/")
	if path == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	store.mu.Lock()
	longURL, exists := store.store[path]
	store.mu.Unlock()
	if !exists {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	res.Header().Set("Location", longURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", handlerGet)
	mux.HandleFunc("/", handlerPost)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed: ", err)
	}
}

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	t.Run("valid ID generation", func(t *testing.T) {
		id, err := generateID()
		assert.NoError(t, err)
		assert.Len(t, id, 13)

		_, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(id)
		assert.NoError(t, err)
	})

	t.Run("uniqueness check", func(t *testing.T) {
		const iterations = 1000
		ids := make(map[string]struct{}, iterations)
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				id, err := generateID()
				assert.NoError(t, err)

				mu.Lock()
				defer mu.Unlock()
				_, exists := ids[id]
				assert.False(t, exists)
				ids[id] = struct{}{}
			}()
		}
		wg.Wait()
	})

	t.Run("error handling", func(t *testing.T) {
		originalReader := rand.Reader
		t.Cleanup(func() { rand.Reader = originalReader })

		rand.Reader = &faultyReader{}
		_, err := generateID()
		assert.Error(t, err)
	})
}

func TestHandlerPost(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name         string
		method       string
		contentType  string
		body         string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "valid request",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "https://example.com",
			wantStatus:   http.StatusCreated,
			wantContains: "http://localhost:8080/",
		},
		{
			name:        "invalid method",
			method:      http.MethodGet,
			contentType: "text/plain",
			body:        "https://example.com",
			wantStatus:  http.StatusMethodNotAllowed,
		},
		{
			name:        "empty body",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        "",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "wrong content type",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"url":"https://example.com"}`,
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.mu.Lock()
			store.store = make(map[string]string)
			store.mu.Unlock()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantContains != "" {
				assert.Contains(t, w.Body.String(), tt.wantContains)
			}

			if tt.wantStatus == http.StatusCreated {
				id := strings.TrimPrefix(w.Body.String(), "http://localhost:8080/")
				store.mu.RLock()
				_, exists := store.store[id]
				store.mu.RUnlock()
				assert.True(t, exists)
			}
		})
	}
}

func TestHandlerGet(t *testing.T) {
	router := setupRouter()
	testID := "testid123"
	testURL := "https://example.org"

	tests := []struct {
		name         string
		url          string
		method       string
		wantStatus   int
		wantLocation string
	}{
		{
			name:         "valid redirect",
			url:          "/" + testID,
			method:       http.MethodGet,
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: testURL,
		},
		{
			name:       "not found",
			url:        "/invalidid",
			method:     http.MethodGet,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid method",
			url:        "/" + testID,
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "empty id",
			url:        "/",
			method:     http.MethodGet,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.mu.Lock()
			store.store = map[string]string{testID: testURL}
			store.mu.Unlock()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.url, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, w.Header().Get("Location"))
			}
		})
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", handlerPost)
	router.GET("/:id", handlerGet)
	return router
}

type faultyReader struct{}

func (r *faultyReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func TestMain(m *testing.M) {
	store = urlStore{
		store: make(map[string]string),
	}
	m.Run()
}

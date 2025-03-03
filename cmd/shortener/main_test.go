package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", handlerPost)
	router.GET("/:id", handlerGet)
	return router
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
			wantStatus:  http.StatusNotFound,
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
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantContains != "" {
				assert.Contains(t, w.Body.String(), tt.wantContains)
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
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty id",
			url:        "/",
			method:     http.MethodGet,
			wantStatus: http.StatusNotFound,
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

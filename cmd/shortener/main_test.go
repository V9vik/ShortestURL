package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGenerateID(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Successful generation",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateID()
			if (err != nil) != tt.wantErr {
				t.Errorf("generateID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 13 {
				t.Errorf("generateID() returned invalid length, got = %d, want = 13", len(got))
			}
		})
	}
}
func TestHandlerPost(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		body         string
		statusCode   int
		responseBody string
	}{
		{
			name:         "Success with text/plain",
			contentType:  "text/plain",
			body:         "https://example.com",
			statusCode:   http.StatusCreated,
			responseBody: "http://localhost:8080/",
		},
		{
			name:         "Failure with empty body",
			contentType:  "text/plain",
			body:         "",
			statusCode:   http.StatusBadRequest,
			responseBody: `{"error":"Empty body"}`,
		},
		{
			name:         "Failure with wrong content-type",
			contentType:  "application/json",
			body:         `{"url": "https://example.com"}`,
			statusCode:   http.StatusBadRequest,
			responseBody: `{"error":"Invalid content type"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/", func(c *gin.Context) {
				handlerPost(c, "http://localhost:8080/")
			})

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status code %d, but got %d", tt.statusCode, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.responseBody) {
				t.Errorf("Response body does not contain expected value:\nGot: %s\nWant: %s", w.Body.String(), tt.responseBody)
			}
		})
	}
}

func TestHandlerGet(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		statusCode   int
		location     string
		errorMessage string
	}{
		{
			name:         "Existing ID",
			id:           "existing-id",
			statusCode:   http.StatusTemporaryRedirect,
			location:     "https://example.com",
			errorMessage: "",
		},
		{
			name:       "Non-existing ID",
			id:         "non-existing-id",
			statusCode: http.StatusTemporaryRedirect,
			location:   "",
		},
		{
			name:       "Empty ID",
			id:         "",
			statusCode: http.StatusNotFound,
			location:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.GET("/:id", func(c *gin.Context) {
				handlerGet(c)
			})

			store.mu.Lock()
			store.store[tt.id] = tt.location
			store.mu.Unlock()

			req, _ := http.NewRequest(http.MethodGet, "/"+tt.id, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status code %d, but got %d", tt.statusCode, w.Code)
			}

			if tt.errorMessage != "" && !strings.Contains(w.Body.String(), tt.errorMessage) {
				t.Errorf("Response body does not contain expected error message:\nGot: %s\nWant: %s", w.Body.String(), tt.errorMessage)
			}

			if tt.location != "" {
				if location := w.Result().Header.Get("Location"); location != tt.location {
					t.Errorf("Expected Location header %s, but got %s", tt.location, location)
				}
			}
		})
	}
}

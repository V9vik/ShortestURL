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
)

func TestGenerateID(t *testing.T) {
	t.Run("valid ID generation", func(t *testing.T) {
		id, err := generateID()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(id) != 13 {
			t.Errorf("Expected ID length 13, got %d", len(id))
		}

		encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
		if _, err := encoder.DecodeString(id); err != nil {
			t.Errorf("Invalid base32 string: %v", err)
		}
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
				if err != nil {
					t.Errorf("Generation error: %v", err)
					return
				}

				mu.Lock()
				defer mu.Unlock()
				if _, exists := ids[id]; exists {
					t.Errorf("Duplicate ID found: %s", id)
				}
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
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestHandlerPost(t *testing.T) {
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
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handlerPost(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantContains != "" && !strings.Contains(rr.Body.String(), tt.wantContains) {
				t.Errorf("Body = %v, should contain %v", rr.Body.String(), tt.wantContains)
			}

			if tt.wantStatus == http.StatusCreated {
				id := strings.TrimPrefix(rr.Body.String(), "http://localhost:8080/")
				store.mu.RLock()
				defer store.mu.RUnlock()
				if _, exists := store.store[id]; !exists {
					t.Errorf("Generated ID %s not found in storage", id)
				}
			}
		})
	}
}

func TestHandlerGet(t *testing.T) {
	testID := "testid123"
	testURL := "https://example.org"
	store.mu.Lock()
	store.store[testID] = testURL
	store.mu.Unlock()

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
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			handlerGet(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", rr.Code, tt.wantStatus)
			}

			if location := rr.Header().Get("Location"); location != tt.wantLocation {
				t.Errorf("Location header = %v, want %v", location, tt.wantLocation)
			}
		})
	}
}

type faultyReader struct{}

func (r *faultyReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}
